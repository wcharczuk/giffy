package cron

// NOTE: ALL TIMES ARE IN UTC. JUST USE UTC.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blend/go-sdk/collections"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// New returns a new job manager.
func New() *JobManager {
	jm := JobManager{
		heartbeatInterval:     DefaultHeartbeatInterval,
		loadedJobs:            map[string]Job{},
		runningTasks:          map[string]Task{},
		schedules:             map[string]Schedule{},
		contexts:              map[string]context.Context{},
		cancels:               map[string]context.CancelFunc{},
		runningTaskStartTimes: map[string]time.Time{},
		lastRunTimes:          map[string]time.Time{},
		nextRunTimes:          map[string]*time.Time{},
		disabledJobs:          collections.SetOfString{},
		enabledProviders:      map[string]func() bool{},
	}

	return &jm
}

// NewFromConfig returns a new job manager from a given config.
func NewFromConfig(cfg *Config) *JobManager {
	return New().WithHeartbeatInterval(cfg.GetHeartbeatInterval())
}

// NewFromEnv returns a new job manager from the environment.
func NewFromEnv() *JobManager {
	return NewFromConfig(NewConfigFromEnv())
}

// JobManager is the main orchestration and job management object.
type JobManager struct {
	heartbeatInterval time.Duration

	loadedJobsLock sync.Mutex
	loadedJobs     map[string]Job

	disabledJobsLock sync.Mutex
	disabledJobs     collections.SetOfString

	enabledProvidersLock sync.Mutex
	enabledProviders     map[string]func() bool

	runningTasksLock sync.Mutex
	runningTasks     map[string]Task

	schedulesLock sync.Mutex
	schedules     map[string]Schedule

	runningTaskStartTimesLock sync.Mutex
	runningTaskStartTimes     map[string]time.Time

	contextsLock sync.Mutex
	contexts     map[string]context.Context

	cancelsLock sync.Mutex
	cancels     map[string]context.CancelFunc

	lastRunTimesLock sync.Mutex
	lastRunTimes     map[string]time.Time

	nextRunTimesLock sync.Mutex
	nextRunTimes     map[string]*time.Time

	schedulerCancel context.CancelFunc
	isRunning       bool

	log *logger.Logger
}

// Logger returns the diagnostics agent.
func (jm *JobManager) Logger() *logger.Logger {
	return jm.log
}

// WithLogger sets the logger and returns a reference to the job manager.
func (jm *JobManager) WithLogger(log *logger.Logger) *JobManager {
	jm.SetLogger(log)
	return jm
}

// SetLogger sets the logger.
func (jm *JobManager) SetLogger(log *logger.Logger) {
	jm.log = log
}

// WithHighPrecisionHeartbeat sets the heartbeat interval to the high precision interval and returns a reference.
func (jm *JobManager) WithHighPrecisionHeartbeat() *JobManager {
	jm.heartbeatInterval = DefaultHighPrecisionHeartbeatInterval
	return jm
}

// WithDefaultPrecisionHeartbeat sets the heartbeat interval to the high precision interval and returns a reference.
func (jm *JobManager) WithDefaultPrecisionHeartbeat() *JobManager {
	jm.heartbeatInterval = DefaultHeartbeatInterval
	return jm
}

// WithHeartbeatInterval sets the heartbeat interval explicitly and returns the job manager.
func (jm *JobManager) WithHeartbeatInterval(interval time.Duration) *JobManager {
	jm.SetHeartbeatInterval(interval)
	return jm
}

// SetHeartbeatInterval sets the heartbeat interval explicitly.
func (jm *JobManager) SetHeartbeatInterval(interval time.Duration) {
	jm.heartbeatInterval = interval
}

// HeartbeatInterval returns the current heartbeat interval.
func (jm *JobManager) HeartbeatInterval() time.Duration {
	return jm.heartbeatInterval
}

// ShouldTriggerListeners is a helper function to determine if we should trigger listeners for a given task.
func (jm *JobManager) ShouldTriggerListeners(taskName string) bool {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	if job, hasJob := jm.loadedJobs[taskName]; hasJob {
		if typed, isTyped := job.(EventTriggerListenersProvider); isTyped {
			return typed.ShouldTriggerListeners()
		}
	}

	return true
}

// ShouldWriteOutput is a helper function to determine if we should write logging output for a task.
func (jm *JobManager) ShouldWriteOutput(taskName string) bool {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	if job, hasJob := jm.loadedJobs[taskName]; hasJob {
		if typed, isTyped := job.(EventShouldWriteOutputProvider); isTyped {
			return typed.ShouldWriteOutput()
		}
	}
	return true
}

// fireTaskStartedListeners fires the currently configured task listeners.
func (jm *JobManager) fireTaskStartedListeners(taskName string) {
	if jm.log == nil {
		return
	}
	jm.log.Trigger(NewEvent(FlagStarted, taskName).WithIsEnabled(jm.ShouldTriggerListeners(taskName)).WithIsWritable(jm.ShouldWriteOutput(taskName)))
}

// fireTaskListeners fires the currently configured task listeners.
func (jm *JobManager) fireTaskCompleteListeners(taskName string, elapsed time.Duration, err error) {
	if jm.log == nil {
		return
	}

	jm.log.Trigger(NewEvent(FlagComplete, taskName).
		WithIsEnabled(jm.ShouldTriggerListeners(taskName)).
		WithIsWritable(jm.ShouldWriteOutput(taskName)).
		WithElapsed(elapsed).
		WithErr(err))

	if err != nil {
		jm.log.Error(err)
	}
}

// shouldRunJob returns whether it is legal to run a job based off of a job's attributes and status.
// Use this function to set logic for whether a job should run
func (jm *JobManager) shouldRunJob(job Job) bool {
	return !jm.IsDisabled(job.Name())
}

// ----------------------------------------------------------------------------
// Informational Methods
// ----------------------------------------------------------------------------

// HasJob returns if a jobName is loaded or not.
func (jm *JobManager) HasJob(jobName string) bool {
	jm.loadedJobsLock.Lock()
	_, hasJob := jm.loadedJobs[jobName]
	jm.loadedJobsLock.Unlock()
	return hasJob
}

// Job returns a job instance by name.
func (jm *JobManager) Job(jobName string) (job Job) {
	jm.loadedJobsLock.Lock()
	job = jm.loadedJobs[jobName]
	jm.loadedJobsLock.Unlock()
	return
}

// IsDisabled returns if a job is disabled.
func (jm *JobManager) IsDisabled(jobName string) (value bool) {
	jm.disabledJobsLock.Lock()
	value = jm.disabledJobs.Contains(jobName)
	jm.disabledJobsLock.Unlock()

	jm.enabledProvidersLock.Lock()
	if provider, hasProvider := jm.enabledProviders[jobName]; hasProvider {
		value = value || !provider()
	}
	jm.enabledProvidersLock.Unlock()
	return
}

// IsRunning returns if a task is currently running.
func (jm *JobManager) IsRunning(taskName string) bool {
	jm.runningTasksLock.Lock()
	_, isRunning := jm.runningTasks[taskName]
	jm.runningTasksLock.Unlock()
	return isRunning
}

// ReadAllJobs allows the consumer to do something with the full job list, using a read lock.
func (jm *JobManager) ReadAllJobs(action func(jobs map[string]Job)) {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()
	action(jm.loadedJobs)
}

// --------------------------------------------------------------------------------
// Core Methods
// --------------------------------------------------------------------------------

// LoadJobs loads a variadic list of jobs.
func (jm *JobManager) LoadJobs(jobs ...Job) error {
	var err error
	for _, job := range jobs {
		err = jm.LoadJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadJob adds a job to the manager.
func (jm *JobManager) LoadJob(j Job) error {
	jobName := j.Name()

	if jm.HasJob(jobName) {
		return exception.NewFromErr(ErrJobAlreadyLoaded).WithMessagef("job: %s", j.Name())
	}

	jm.setLoadedJob(jobName, j)
	jobSchedule := j.Schedule()
	jm.setSchedule(jobName, jobSchedule)
	jm.setNextRunTime(jobName, jobSchedule.GetNextRunTime(nil))
	jm.setEnabledProvider(jobName, j)
	return nil
}

// DisableJobs disables a variadic list of job names.
func (jm *JobManager) DisableJobs(jobNames ...string) error {
	var err error
	for _, jobName := range jobNames {
		err = jm.DisableJob(jobName)
		if err != nil {
			return err
		}
	}
	return nil
}

// DisableJob stops a job from running but does not unload it.
func (jm *JobManager) DisableJob(jobName string) error {
	if !jm.HasJob(jobName) {
		return exception.NewFromErr(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
	}

	jm.setDisabledJob(jobName)
	jm.deleteNextRunTime(jobName)
	return nil
}

// EnableJobs enables a variadic list of job names.
func (jm *JobManager) EnableJobs(jobNames ...string) error {
	var err error
	for _, jobName := range jobNames {
		err = jm.EnableJob(jobName)
		if err != nil {
			return err
		}
	}
	return nil
}

// EnableJob enables a job that has been disabled.
func (jm *JobManager) EnableJob(jobName string) error {
	if !jm.HasJob(jobName) {
		return exception.NewFromErr(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
	}

	jm.deleteDisabledJob(jobName)
	job := jm.getLoadedJob(jobName)
	jobSchedule := job.Schedule()
	jm.setNextRunTime(jobName, jobSchedule.GetNextRunTime(nil))

	return nil
}

// RunJobs runs a variadic list of job names.
func (jm *JobManager) RunJobs(jobNames ...string) error {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	for _, jobName := range jobNames {
		if job, hasJob := jm.loadedJobs[jobName]; hasJob {
			if !jm.IsDisabled(jobName) {
				jobErr := jm.RunTask(job)
				if jobErr != nil {
					return jobErr
				}
			}
		} else {
			return exception.NewFromErr(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
		}
	}
	return nil
}

// RunJob runs a job by jobName on demand.
func (jm *JobManager) RunJob(jobName string) error {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	if job, hasJob := jm.loadedJobs[jobName]; hasJob {
		if jm.shouldRunJob(job) {
			now := time.Now().UTC()
			jm.setLastRunTime(jobName, now)
			err := jm.RunTask(job)
			return err
		}
		return nil
	}
	return exception.NewFromErr(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
}

// RunAllJobs runs every job that has been loaded in the JobManager at once.
func (jm *JobManager) RunAllJobs() error {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	for jobName, job := range jm.loadedJobs {
		if !jm.IsDisabled(jobName) {
			jobErr := jm.RunTask(job)
			if jobErr != nil {
				return jobErr
			}
		}
	}
	return nil
}

// RunTask runs a task on demand.
func (jm *JobManager) RunTask(t Task) error {
	if !jm.shouldRunTask(t) {
		return nil
	}
	taskName := t.Name()
	start := Now()
	ctx, cancel := jm.createContext()

	jm.setRunningTask(taskName, t)
	jm.setContext(ctx, taskName)
	jm.setCancelFunc(taskName, cancel)
	jm.setRunningTaskStartTime(taskName, start)
	jm.setLastRunTime(taskName, start)

	// this is the main goroutine that runs the task
	go func() {
		var err error

		defer func() {
			jm.cleanupTask(taskName)
			jm.fireTaskCompleteListeners(taskName, Since(start), err)
		}()

		// panic recovery
		defer func() {
			if r := recover(); r != nil {
				err = exception.Newf("%v", r)
			}
		}()

		jm.onTaskStart(t)
		jm.fireTaskStartedListeners(taskName)
		err = t.Execute(ctx)
		jm.onTaskComplete(t, err)
	}()

	return nil
}

// CancelTask cancels (sends the cancellation signal) to a running task.
func (jm *JobManager) CancelTask(taskName string) error {
	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	if task, hasTask := jm.runningTasks[taskName]; hasTask {
		cancel := jm.getCancelFunc(taskName)
		jm.onTaskCancellation(task)
		cancel()
	}

	return exception.NewFromErr(ErrTaskNotFound).WithMessagef("task: %s", taskName)
}

// Start begins the schedule runner for a JobManager.
func (jm *JobManager) Start() {
	ctx, cancel := jm.createContext()
	jm.schedulerCancel = cancel

	go jm.runDueJobs(ctx)
	go jm.killHangingJobs(ctx)

	jm.isRunning = true
}

// Stop stops the schedule runner for a JobManager.
func (jm *JobManager) Stop() {
	if !jm.isRunning {
		return
	}
	jm.schedulerCancel()
	jm.isRunning = false
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

func (jm *JobManager) createContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// shouldRunTask returns whether a task should be executed based on its status
func (jm *JobManager) shouldRunTask(t Task) bool {
	_, serial := t.(SerialProvider)
	if serial {
		return !jm.IsRunning(t.Name())
	}
	return true
}

func (jm *JobManager) onTaskStart(t Task) {
	if receiver, isReceiver := t.(OnStartReceiver); isReceiver {
		receiver.OnStart()
	}
}

func (jm *JobManager) onTaskComplete(t Task, result error) {
	if receiver, isReceiver := t.(OnCompleteReceiver); isReceiver {
		receiver.OnComplete(result)
	}
}

func (jm *JobManager) onTaskCancellation(t Task) {
	if receiver, isReceiver := t.(OnCancellationReceiver); isReceiver {
		receiver.OnCancellation()
	}
}

func (jm *JobManager) cleanupTask(taskName string) {
	jm.deleteRunningTaskStartTime(taskName)
	jm.deleteRunningTask(taskName)
	jm.deleteContext(taskName)
	jm.deleteCancelFunc(taskName)
}

func (jm *JobManager) runDueJobs(ctx context.Context) {
	heartbeat := time.Tick(jm.heartbeatInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat:
			jm.runDueJobsInner()
		}
	}
}

func (jm *JobManager) runDueJobsInner() {
	now := Now()
	var err error
	for jobName, job := range jm.loadedJobs {
		nextRunTime := jm.getNextRunTime(jobName)
		if nextRunTime != nil {
			if jm.shouldRunJob(job) {
				if nextRunTime.Before(now) {
					jm.setNextRunTime(jobName, jm.getSchedule(jobName).GetNextRunTime(&now))
					jm.setLastRunTime(jobName, now)
					err = jm.RunTask(job)
					if err != nil {
						jm.log.Error(err)
					}
				}
			}
		}
	}
}

func (jm *JobManager) killHangingJobs(ctx context.Context) {
	heartbeat := time.Tick(jm.heartbeatInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat:
			jm.killHangingJobsInner()
		}
	}
}

func (jm *JobManager) killHangingJobsInner() error {
	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	jm.cancelsLock.Lock()
	defer jm.cancelsLock.Unlock()

	jm.nextRunTimesLock.Lock()
	defer jm.nextRunTimesLock.Unlock()

	var err error
	for taskName, startedTime := range jm.runningTaskStartTimes {
		task, hasTask := jm.runningTasks[taskName]
		if !hasTask {
			continue
		}

		timeoutProvider, isTimeoutProvder := task.(TimeoutProvider)
		if !isTimeoutProvder {
			continue
		}
		timeout := timeoutProvider.Timeout()
		currentTime := Now()
		if nextRunTime, hasNextRuntime := jm.nextRunTimes[taskName]; hasNextRuntime {
			// we need to calculate the effective timeout
			// either startedTime+timeout or the next runtime, whichever is closer.

			// t1 represents the absolute timeout time.
			t1 := startedTime.Add(timeout)
			// t2 represents the next runtime, or an effective time we need to stop by.
			t2 := *nextRunTime

			// the effective timeout is whichever is more soon.
			effectiveTimeout := Min(t1, t2)

			// if the effective timeout is in the past, or it's within the next heartbeat.
			if currentTime.After(effectiveTimeout) || effectiveTimeout.Sub(currentTime) < jm.heartbeatInterval {
				err = jm.killHangingJob(taskName)
				if err != nil {
					jm.log.Error(err)
				}
			}
		} else if currentTime.Sub(startedTime) >= timeout {
			err = jm.killHangingJob(taskName)
			if err != nil {
				jm.log.Error(err)
			}
		}
	}
	return nil
}

// killHangingJob cancels (sends the cancellation signal) to a running task that has exceeded its timeout.
// it assumes that the following locks are held:
// - runningTasksLock
// - runningTaskStartTimesLock
// - contextsLock
// otherwise, chaos, mayhem, deadlocks. You should *rarely* need to call this explicitly.
func (jm *JobManager) killHangingJob(taskName string) error {
	_, hasTask := jm.runningTasks[taskName]
	if !hasTask {
		return exception.Newf("task not found").WithMessagef("Task: %s", taskName)
	}

	cancel, hasCancel := jm.cancels[taskName]
	if !hasCancel {
		return exception.Newf("task has no cancellation handle").WithMessagef("Task: %s", taskName)
	}

	cancel()

	delete(jm.runningTasks, taskName)
	delete(jm.runningTaskStartTimes, taskName)
	delete(jm.contexts, taskName)
	delete(jm.cancels, taskName)

	return nil
}

// --------------------------------------------------------------------------------
// Status Methods
// --------------------------------------------------------------------------------

// Status returns the status metadata for a JobManager
func (jm *JobManager) Status() []TaskStatus {

	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	jm.disabledJobsLock.Lock()
	defer jm.disabledJobsLock.Unlock()

	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	jm.nextRunTimesLock.Lock()
	defer func() {
		jm.nextRunTimesLock.Unlock()
	}()

	jm.lastRunTimesLock.Lock()
	defer jm.lastRunTimesLock.Unlock()

	var statuses []TaskStatus
	now := Now()
	for jobName, job := range jm.loadedJobs {
		status := TaskStatus{}
		status.Name = jobName

		if runningSince, isRunning := jm.runningTaskStartTimes[jobName]; isRunning {
			status.State = StateRunning
			status.RunningFor = fmt.Sprintf("%v", now.Sub(runningSince))
		} else if jm.disabledJobs.Contains(jobName) {
			status.State = StateDisabled
		} else {
			status.State = StateEnabled
		}

		if lastRunTime, hasLastRunTime := jm.lastRunTimes[jobName]; hasLastRunTime {
			status.LastRunTime = FormatTime(lastRunTime)
		}

		if nextRunTime, hasNextRunTime := jm.nextRunTimes[jobName]; hasNextRunTime {
			if nextRunTime != nil {
				status.NextRunTime = FormatTime(*nextRunTime)
			}
		}

		if statusProvider, isStatusProvider := job.(StatusProvider); isStatusProvider {
			providedStatus := statusProvider.Status()
			if len(providedStatus) > 0 {
				status.Status = providedStatus
			}
		}

		statuses = append(statuses, status)
	}

	for taskName, task := range jm.runningTasks {
		if _, isJob := jm.loadedJobs[taskName]; !isJob {
			status := TaskStatus{
				Name:  taskName,
				State: StateRunning,
			}
			if runningSince, isRunning := jm.runningTaskStartTimes[taskName]; isRunning {
				status.RunningFor = fmt.Sprintf("%v", now.Sub(runningSince))
			}
			if statusProvider, isStatusProvider := task.(StatusProvider); isStatusProvider {
				status.Status = statusProvider.Status()
			}
			statuses = append(statuses, status)
		}
	}
	return statuses
}

// TaskStatus returns the status metadata for a given task.
func (jm *JobManager) TaskStatus(taskName string) *TaskStatus {
	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	if task, isRunning := jm.runningTasks[taskName]; isRunning {
		now := Now()
		status := TaskStatus{
			Name:  taskName,
			State: StateRunning,
		}
		if runningSince, isRunning := jm.runningTaskStartTimes[taskName]; isRunning {
			status.RunningFor = fmt.Sprintf("%v", now.Sub(runningSince))
		}
		if statusProvider, isStatusProvider := task.(StatusProvider); isStatusProvider {
			status.Status = statusProvider.Status()
		}
		return &status
	}
	return nil
}

// --------------------------------------------------------------------------------
// Atomic Access/Mutating Methods
// --------------------------------------------------------------------------------

func (jm *JobManager) getContext(taskName string) (ctx context.Context) {
	jm.contextsLock.Lock()
	ctx = jm.contexts[taskName]
	jm.contextsLock.Unlock()
	return
}

// note; this setter is out of order because of the linter.
func (jm *JobManager) setContext(ctx context.Context, taskName string) {
	jm.contextsLock.Lock()
	jm.contexts[taskName] = ctx
	jm.contextsLock.Unlock()
}

func (jm *JobManager) deleteContext(taskName string) {
	jm.contextsLock.Lock()
	delete(jm.contexts, taskName)
	jm.contextsLock.Unlock()
}

func (jm *JobManager) getCancelFunc(taskName string) (cancel context.CancelFunc) {
	jm.cancelsLock.Lock()
	cancel = jm.cancels[taskName]
	jm.cancelsLock.Unlock()
	return
}

func (jm *JobManager) setCancelFunc(taskName string, cancel context.CancelFunc) {
	jm.cancelsLock.Lock()
	jm.cancels[taskName] = cancel
	jm.cancelsLock.Unlock()
}

func (jm *JobManager) deleteCancelFunc(taskName string) {
	jm.cancelsLock.Lock()
	delete(jm.cancels, taskName)
	jm.cancelsLock.Unlock()
}

func (jm *JobManager) setDisabledJob(jobName string) {
	jm.disabledJobsLock.Lock()
	jm.disabledJobs.Add(jobName)
	jm.disabledJobsLock.Unlock()
}

func (jm *JobManager) deleteDisabledJob(jobName string) {
	jm.disabledJobsLock.Lock()
	jm.disabledJobs.Remove(jobName)
	jm.disabledJobsLock.Unlock()
}

func (jm *JobManager) getNextRunTime(jobName string) (nextRunTime *time.Time) {
	jm.nextRunTimesLock.Lock()
	nextRunTime = jm.nextRunTimes[jobName]
	jm.nextRunTimesLock.Unlock()
	return
}

func (jm *JobManager) setNextRunTime(jobName string, t *time.Time) {
	jm.nextRunTimesLock.Lock()
	jm.nextRunTimes[jobName] = t
	jm.nextRunTimesLock.Unlock()
}

func (jm *JobManager) setEnabledProvider(jobName string, j Job) {

	if typed, isTyped := j.(EnabledProvider); isTyped {
		jm.enabledProvidersLock.Lock()
		jm.enabledProviders[jobName] = func() bool { return typed.Enabled() }
		jm.enabledProvidersLock.Unlock()
	}
}

func (jm *JobManager) deleteNextRunTime(jobName string) {
	jm.nextRunTimesLock.Lock()
	delete(jm.nextRunTimes, jobName)
	jm.nextRunTimesLock.Unlock()
}

func (jm *JobManager) getLastRunTime(taskName string) (lastRunTime time.Time) {
	jm.lastRunTimesLock.Lock()
	lastRunTime = jm.lastRunTimes[taskName]
	jm.lastRunTimesLock.Unlock()
	return
}

func (jm *JobManager) setLastRunTime(taskName string, t time.Time) {
	jm.lastRunTimesLock.Lock()
	jm.lastRunTimes[taskName] = t
	jm.lastRunTimesLock.Unlock()
}

func (jm *JobManager) deleteLastRunTime(taskName string) {
	jm.lastRunTimesLock.Lock()
	delete(jm.lastRunTimes, taskName)
	jm.lastRunTimesLock.Unlock()
}

func (jm *JobManager) getLoadedJob(jobName string) (job Job) {
	jm.loadedJobsLock.Lock()
	job = jm.loadedJobs[jobName]
	jm.loadedJobsLock.Unlock()
	return
}

func (jm *JobManager) setLoadedJob(jobName string, j Job) {
	jm.loadedJobsLock.Lock()
	jm.loadedJobs[jobName] = j
	jm.loadedJobsLock.Unlock()
}

func (jm *JobManager) deleteLoadedJob(jobName string) {
	jm.loadedJobsLock.Lock()
	delete(jm.loadedJobs, jobName)
	jm.loadedJobsLock.Unlock()
}

func (jm *JobManager) getRunningTask(taskName string) (task Task) {
	jm.runningTasksLock.Lock()
	task = jm.runningTasks[taskName]
	jm.runningTasksLock.Unlock()
	return
}

func (jm *JobManager) setRunningTask(taskName string, t Task) {
	jm.runningTasksLock.Lock()
	jm.runningTasks[taskName] = t
	jm.runningTasksLock.Unlock()
}

func (jm *JobManager) deleteRunningTask(taskName string) {
	jm.runningTasksLock.Lock()
	delete(jm.runningTasks, taskName)
	jm.runningTasksLock.Unlock()
}

func (jm *JobManager) getRunningTaskStartTime(taskName string) (startTime time.Time) {
	jm.runningTaskStartTimesLock.Lock()

	startTime = jm.runningTaskStartTimes[taskName]
	jm.runningTaskStartTimesLock.Unlock()
	return
}

func (jm *JobManager) setRunningTaskStartTime(taskName string, t time.Time) {
	jm.runningTaskStartTimesLock.Lock()
	jm.runningTaskStartTimes[taskName] = t
	jm.runningTaskStartTimesLock.Unlock()
}

func (jm *JobManager) deleteRunningTaskStartTime(taskName string) {
	jm.runningTaskStartTimesLock.Lock()
	delete(jm.runningTaskStartTimes, taskName)
	jm.runningTaskStartTimesLock.Unlock()
}

func (jm *JobManager) getSchedule(jobName string) (schedule Schedule) {
	jm.schedulesLock.Lock()
	schedule = jm.schedules[jobName]
	jm.schedulesLock.Unlock()
	return
}

func (jm *JobManager) setSchedule(jobName string, schedule Schedule) {
	jm.schedulesLock.Lock()
	jm.schedules[jobName] = schedule
	jm.schedulesLock.Unlock()
}

func (jm *JobManager) deleteSchedule(jobName string) {
	jm.schedulesLock.Lock()
	defer jm.schedulesLock.Unlock()

	delete(jm.schedules, jobName)
}
