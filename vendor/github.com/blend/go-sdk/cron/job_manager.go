/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package cron

// NOTE: ALL TIMES ARE IN UTC. JUST USE UTC.

import (
	"context"
	"sync"
	"time"

	"github.com/blend/go-sdk/async"
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
)

// New returns a new job manager.
func New(options ...JobManagerOption) *JobManager {
	jm := JobManager{
		Latch:       async.NewLatch(),
		BaseContext: context.Background(),
		Jobs:        make(map[string]*JobScheduler),
	}
	for _, option := range options {
		option(&jm)
	}
	return &jm
}

// JobManager is the main orchestration and job management object.
type JobManager struct {
	sync.Mutex
	Latch       *async.Latch
	BaseContext context.Context
	Tracer      Tracer
	Log         logger.Log
	Started     time.Time
	Stopped     time.Time
	Jobs        map[string]*JobScheduler
}

//
// Life Cycle
//

// Start starts the job manager and blocks.
func (jm *JobManager) Start() error {
	if err := jm.StartAsync(); err != nil {
		return err
	}
	<-jm.Latch.NotifyStopped()
	return nil
}

// StartAsync starts the job manager and the loaded jobs.
// It does not block.
func (jm *JobManager) StartAsync() error {
	if !jm.Latch.CanStart() {
		return async.ErrCannotStart
	}
	jm.Latch.Starting()
	logger.MaybeInfo(jm.Log, "job manager starting")
	for _, jobScheduler := range jm.Jobs {
		errors := make(chan error)
		go func() {
			errors <- jobScheduler.Start()
		}()
		logger.MaybeDebugf(jm.Log, "job manager starting job %s", jobScheduler.Name())
		select {
		case err := <-errors:
			logger.MaybeError(jm.Log, err)
		case <-jobScheduler.NotifyStarted():
			logger.MaybeDebugf(jm.Log, "job manager starting job %s complete", jobScheduler.Name())
			continue
		}
	}

	jm.Latch.Started()
	jm.Started = time.Now().UTC()
	logger.MaybeInfo(jm.Log, "job manager started")
	return nil
}

// Stop stops the schedule runner for a JobManager.
func (jm *JobManager) Stop() error {
	if !jm.Latch.CanStop() {
		return async.ErrCannotStop
	}
	jm.Latch.Stopping()
	logger.MaybeInfo(jm.Log, "job manager stopping")
	defer func() {
		jm.Stopped = time.Now().UTC()
		jm.Latch.Stopped()
		jm.Latch.Reset()
		logger.MaybeInfo(jm.Log, "job manager stopping complete")
	}()
	for _, jobScheduler := range jm.Jobs {
		if err := jobScheduler.OnUnload(jobScheduler.Background()); err != nil {
			logger.MaybeError(jm.Log, err)
		}
		if err := jobScheduler.Stop(); err != nil {
			logger.MaybeError(jm.Log, err)
		}
	}
	return nil
}

//
// job management
//

// LoadJobs loads a variadic list of jobs.
func (jm *JobManager) LoadJobs(jobs ...Job) error {
	jm.Lock()
	defer jm.Unlock()

	for _, job := range jobs {
		jobName := job.Name()
		if _, hasJob := jm.Jobs[jobName]; hasJob {
			return ex.New(ErrJobAlreadyLoaded, ex.OptMessagef("job: %s", job.Name()))
		}

		jobScheduler := NewJobScheduler(
			job,
			OptJobSchedulerLog(jm.Log),
			OptJobSchedulerTracer(jm.Tracer),
			OptJobSchedulerBaseContext(jm.BaseContext),
		)
		if err := jobScheduler.OnLoad(jobScheduler.Background()); err != nil {
			return err
		}
		jm.Jobs[jobName] = jobScheduler
	}
	return nil
}

// UnloadJobs removes jobs from the manager and stops them.
func (jm *JobManager) UnloadJobs(jobNames ...string) error {
	jm.Lock()
	defer jm.Unlock()

	for _, jobName := range jobNames {
		if jobScheduler, ok := jm.Jobs[jobName]; ok {
			if err := jobScheduler.OnUnload(jobScheduler.Background()); err != nil {
				return err
			}
			if err := jobScheduler.Stop(); err != nil {
				logger.MaybeError(jm.Log, err)
			}
			delete(jm.Jobs, jobName)
		} else {
			return ex.New(ErrJobNotFound, ex.OptMessagef("job: %s", jobName))
		}
	}
	return nil
}

// DisableJobs disables a variadic list of job names.
func (jm *JobManager) DisableJobs(jobNames ...string) error {
	jm.Lock()
	defer jm.Unlock()

	for _, jobName := range jobNames {
		if job, ok := jm.Jobs[jobName]; ok {
			job.Disable()
		} else {
			return ex.New(ErrJobNotFound, ex.OptMessagef("job: %s", jobName))
		}
	}
	return nil
}

// EnableJobs enables a variadic list of job names.
func (jm *JobManager) EnableJobs(jobNames ...string) error {
	jm.Lock()
	defer jm.Unlock()

	for _, jobName := range jobNames {
		if job, ok := jm.Jobs[jobName]; ok {
			job.Enable()
		} else {
			return ex.New(ErrJobNotFound, ex.OptMessagef("job: %s", jobName))
		}
	}
	return nil
}

// HasJob returns if a jobName is loaded or not.
func (jm *JobManager) HasJob(jobName string) (hasJob bool) {
	jm.Lock()
	defer jm.Unlock()
	_, hasJob = jm.Jobs[jobName]
	return
}

// Job returns a job metadata by name.
func (jm *JobManager) Job(jobName string) (job *JobScheduler, err error) {
	jm.Lock()
	jobScheduler, hasJob := jm.Jobs[jobName]
	jm.Unlock()
	if hasJob {
		job = jobScheduler
	} else {
		err = ex.New(ErrJobNotLoaded, ex.OptMessagef("job: %s", jobName))
	}
	return
}

// IsJobDisabled returns if a job is disabled.
func (jm *JobManager) IsJobDisabled(jobName string) (value bool) {
	jm.Lock()
	jobScheduler, hasJob := jm.Jobs[jobName]
	jm.Unlock()
	if hasJob {
		value = jobScheduler.Disabled()
	}
	return
}

// IsJobRunning returns if a job is currently running.
func (jm *JobManager) IsJobRunning(jobName string) (isRunning bool) {
	jm.Lock()
	jobScheduler, ok := jm.Jobs[jobName]
	jm.Unlock()
	if ok {
		isRunning = !jobScheduler.IsIdle()
	}
	return
}

// RunJob runs a job by jobName on demand.
func (jm *JobManager) RunJob(jobName string) (*JobInvocation, <-chan struct{}, error) {
	jm.Lock()
	jobScheduler, ok := jm.Jobs[jobName]
	jm.Unlock()
	if !ok {
		return nil, nil, ex.New(ErrJobNotLoaded, ex.OptMessagef("job: %s", jobName))
	}
	return jobScheduler.RunAsync()
}

// RunJobContext runs a job by jobName on demand with a given context.
func (jm *JobManager) RunJobContext(ctx context.Context, jobName string) (*JobInvocation, <-chan struct{}, error) {
	jm.Lock()
	jobScheduler, ok := jm.Jobs[jobName]
	jm.Unlock()
	if !ok {
		return nil, nil, ex.New(ErrJobNotLoaded, ex.OptMessagef("job: %s", jobName))
	}
	return jobScheduler.RunAsyncContext(ctx)
}

// CancelJob cancels (sends the cancellation signal) to a running job.
func (jm *JobManager) CancelJob(jobName string) (err error) {
	jm.Lock()
	jobScheduler, ok := jm.Jobs[jobName]
	jm.Unlock()
	if !ok {
		err = ex.New(ErrJobNotFound, ex.OptMessagef("job: %s", jobName))
		return
	}
	err = jobScheduler.Cancel()
	return
}

//
// status and state
//

// State returns the job manager state.
func (jm *JobManager) State() JobManagerState {
	if jm.Latch.IsStarted() {
		return JobManagerStateRunning
	} else if jm.Latch.IsStopped() {
		return JobManagerStateStopped
	}
	return JobManagerStateUnknown
}
