package chronometer

// NOTE: ALL TIMES ARE IN UTC. JUST USE UTC.

import (
	"fmt"
	"sync"
	"time"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/collections"
)

const (
	// HeartbeatInterval is the interval between schedule next run checks.
	HeartbeatInterval = 250 * time.Millisecond

	// HangingHeartbeatInterval is the interval between schedule next run checks.
	HangingHeartbeatInterval = 333 * time.Millisecond

	//StateRunning is the running state.
	StateRunning = "running"

	// StateEnabled is the enabled state.
	StateEnabled = "enabled"

	// StateDisabled is the disabled state.
	StateDisabled = "disabled"
)

// NewJobManager returns a new instance of JobManager.
func NewJobManager() *JobManager {
	jm := JobManager{
		loadedJobs:            map[string]Job{},
		runningTasks:          map[string]Task{},
		schedules:             map[string]Schedule{},
		cancellationTokens:    map[string]*CancellationToken{},
		runningTaskStartTimes: map[string]time.Time{},
		lastRunTimes:          map[string]time.Time{},
		nextRunTimes:          map[string]*time.Time{},
		disabledJobs:          collections.SetOfString{},
	}

	return &jm
}

var _default *JobManager
var _defaultLock = &sync.Mutex{}

// Default returns a shared instance of a JobManager.
func Default() *JobManager {
	if _default == nil {
		_defaultLock.Lock()
		defer _defaultLock.Unlock()

		if _default == nil {
			_default = NewJobManager()
		}
	}
	return _default
}

// JobManager is the main orchestration and job management object.
type JobManager struct {
	loadedJobsLock sync.RWMutex
	loadedJobs     map[string]Job

	disabledJobsLock sync.RWMutex
	disabledJobs     collections.SetOfString

	runningTasksLock sync.RWMutex
	runningTasks     map[string]Task

	schedulesLock sync.RWMutex
	schedules     map[string]Schedule

	cancellationTokensLock sync.RWMutex
	cancellationTokens     map[string]*CancellationToken

	runningTaskStartTimesLock sync.RWMutex
	runningTaskStartTimes     map[string]time.Time

	lastRunTimesLock sync.RWMutex
	lastRunTimes     map[string]time.Time

	nextRunTimesLock sync.RWMutex
	nextRunTimes     map[string]*time.Time

	isRunning      bool
	schedulerToken *CancellationToken

	logger *logger.Agent
}

// Logger returns the diagnostics agent.
func (jm *JobManager) Logger() *logger.Agent {
	return jm.logger
}

// SetLogger sets the diagnostics agent.
func (jm *JobManager) SetLogger(agent *logger.Agent) {
	jm.logger = agent

	if jm.logger != nil {
		jm.logger.AddEventListener(EventTask, NewTaskListener(jm.taskListener))
		jm.logger.AddEventListener(EventTaskComplete, NewTaskCompleteListener(jm.taskCompleteListener))
	}
}

// ShouldShowMessagesFor is a helper function to determine if we should show messages for a
// given task name.
func (jm *JobManager) ShouldShowMessagesFor(taskName string) bool {
	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()

	if job, hasJob := jm.loadedJobs[taskName]; hasJob {
		if typed, isTyped := job.(ShowMessagesProvider); isTyped {
			return typed.ShowMessages()
		}
	}

	return true
}

func (jm *JobManager) taskListener(wr *logger.Writer, ts logger.TimeSource, taskName string) {
	if jm.ShouldShowMessagesFor(taskName) {
		logger.WriteEventf(wr, ts, EventTask, logger.ColorBlue, "`%s` starting", taskName)
	}
}

func (jm *JobManager) taskCompleteListener(wr *logger.Writer, ts logger.TimeSource, taskName string, elapsed time.Duration, err error) {
	if jm.ShouldShowMessagesFor(taskName) {
		if err != nil {
			logger.WriteEventf(wr, ts, EventTaskComplete, logger.ColorRed, "`%s` failed %v", taskName, elapsed)
		} else {
			logger.WriteEventf(wr, ts, EventTaskComplete, logger.ColorBlue, "`%s` completed %v", taskName, elapsed)
		}
	}
}

// fireTaskListeners fires the currently configured task listeners.
func (jm *JobManager) fireTaskListeners(taskName string) {
	if jm.logger == nil {
		return
	}
	jm.logger.OnEvent(EventTask, taskName)
}

// fireTaskListeners fires the currently configured task listeners.
func (jm *JobManager) fireTaskCompleteListeners(taskName string, elapsed time.Duration, err error) {
	if jm.logger == nil {
		return
	}
	jm.logger.OnEvent(EventTaskComplete, taskName, elapsed, err)
}

// --------------------------------------------------------------------------------
// Informational Methods
// --------------------------------------------------------------------------------

// HasJob returns if a jobName is loaded or not.
func (jm *JobManager) HasJob(jobName string) bool {
	jm.loadedJobsLock.RLock()
	_, hasJob := jm.loadedJobs[jobName]
	jm.loadedJobsLock.RUnlock()
	return hasJob
}

// IsDisabled returns if a job is disabled.
func (jm *JobManager) IsDisabled(jobName string) (value bool) {
	jm.disabledJobsLock.RLock()
	value = jm.disabledJobs.Contains(jobName)
	jm.disabledJobsLock.RUnlock()
	return
}

// IsRunning returns if a task is currently running.
func (jm *JobManager) IsRunning(taskName string) bool {
	jm.runningTasksLock.RLock()
	_, isRunning := jm.runningTasks[taskName]
	jm.runningTasksLock.RUnlock()
	return isRunning
}

// ReadAllJobs allows the consumer to do something with the full job list, using a read lock.
func (jm *JobManager) ReadAllJobs(action func(jobs map[string]Job)) {
	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()
	action(jm.loadedJobs)
}

// --------------------------------------------------------------------------------
// Core Methods
// --------------------------------------------------------------------------------

// LoadJob adds a job to the manager.
func (jm *JobManager) LoadJob(j Job) error {
	jobName := j.Name()

	if jm.HasJob(jobName) {
		return exception.Newf("Job name `%s` already loaded.", j.Name())
	}

	jm.setLoadedJob(jobName, j)
	jobSchedule := j.Schedule()
	jm.setSchedule(jobName, jobSchedule)
	jm.setNextRunTime(jobName, jobSchedule.GetNextRunTime(nil))
	return nil
}

// DisableJob stops a job from running but does not unload it.
func (jm *JobManager) DisableJob(jobName string) error {
	if !jm.HasJob(jobName) {
		return exception.Newf("Job name `%s` isn't loaded.", jobName)
	}

	jm.setDisabledJob(jobName)
	jm.deleteNextRunTime(jobName)
	return nil
}

// EnableJob enables a job that has been disabled.
func (jm *JobManager) EnableJob(jobName string) error {
	if !jm.HasJob(jobName) {
		return exception.Newf("Job name `%s` isn't loaded.", jobName)
	}

	jm.deleteDisabledJob(jobName)
	job := jm.getLoadedJob(jobName)
	jobSchedule := job.Schedule()
	jm.setNextRunTime(jobName, jobSchedule.GetNextRunTime(nil))

	return nil
}

func (jm *JobManager) showJobMessages(job Job) bool {
	hasDiagnostics := jm.logger != nil
	if showMessagesProvider, isShowMessagesProvider := job.(ShowMessagesProvider); isShowMessagesProvider {
		return hasDiagnostics && showMessagesProvider.ShowMessages()
	}
	return hasDiagnostics
}

// RunJob runs a job by jobName on demand.
func (jm *JobManager) RunJob(jobName string) error {
	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()

	jm.disabledJobsLock.RLock()
	defer jm.disabledJobsLock.RUnlock()

	if job, hasJob := jm.loadedJobs[jobName]; hasJob {
		if !jm.disabledJobs.Contains(jobName) {
			now := time.Now().UTC()
			jm.setLastRunTime(jobName, now)
			err := jm.RunTask(job)
			return err
		}
		return nil
	}
	return exception.Newf("Job name `%s` not found.", jobName)
}

// RunAllJobs runs every job that has been loaded in the JobManager at once.
func (jm *JobManager) RunAllJobs() error {
	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()

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
	taskName := t.Name()
	ct := NewCancellationToken()
	start := Now()

	jm.setRunningTask(taskName, t)
	jm.setCancellationToken(taskName, ct)
	jm.setRunningTaskStartTime(taskName, start)
	jm.setLastRunTime(taskName, start)

	// this is the main goroutine that runs the task
	// it itself spawns another goroutine
	go func() {
		var err error
		defer func() {
			jm.cleanupTask(taskName)
			jm.fireTaskCompleteListeners(taskName, Since(start), err)
		}()

		defer func() {
			if r := recover(); r != nil {
				if _, isCancellation := r.(CancellationPanic); isCancellation {
					jm.onTaskCancellation(t)
				}
				if jm.logger != nil {
					jm.logger.Fatalf("%+v", r)
				}
			}
		}()

		jm.onTaskStart(t)
		jm.fireTaskListeners(taskName)
		err = t.Execute(ct)
		jm.onTaskComplete(t, err)
	}()

	return nil
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
	jm.deleteCancellationToken(taskName)
}

// CancelTask cancels (sends the cancellation signal) to a running task.
func (jm *JobManager) CancelTask(taskName string) error {
	jm.runningTasksLock.RLock()
	defer jm.runningTasksLock.RUnlock()

	if task, hasTask := jm.runningTasks[taskName]; hasTask {
		token := jm.getCancellationToken(taskName)
		jm.onTaskCancellation(task)
		token.Cancel()
	}
	return exception.Newf("Task name `%s` not found.", taskName)
}

// Start begins the schedule runner for a JobManager.
func (jm *JobManager) Start() {
	ct := NewCancellationToken()
	jm.schedulerToken = ct

	go jm.runDueJobs(ct)
	go jm.killHangingJobs(ct)

	jm.isRunning = true
}

// Stop stops the schedule runner for a JobManager.
func (jm *JobManager) Stop() {
	if !jm.isRunning {
		return
	}
	jm.schedulerToken.Cancel()
	jm.isRunning = false
}

func (jm *JobManager) runDueJobs(ct *CancellationToken) {
	for !ct.didCancel() {
		jm.runDueJobsInner()
		time.Sleep(HeartbeatInterval)
	}
}

func (jm *JobManager) runDueJobsInner() {
	jm.nextRunTimesLock.Lock()
	defer jm.nextRunTimesLock.Unlock()

	now := Now()

	for jobName, nextRunTime := range jm.nextRunTimes {
		if nextRunTime != nil {
			if !jm.IsDisabled(jobName) {
				if nextRunTime.Before(now) {
					jm.nextRunTimes[jobName] = jm.getSchedule(jobName).GetNextRunTime(&now)
					jm.RunJob(jobName)
				}
			}
		}
	}
}

func (jm *JobManager) killHangingJobs(ct *CancellationToken) {
	for !ct.didCancel() {
		jm.killHangingJobsInner()
		time.Sleep(HangingHeartbeatInterval)
	}
}

func (jm *JobManager) killHangingJobsInner() {

	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	jm.cancellationTokensLock.Lock()
	defer jm.cancellationTokensLock.Unlock()

	now := Now()
	for taskName, startedTime := range jm.runningTaskStartTimes {
		if task, hasTask := jm.runningTasks[taskName]; hasTask {
			if timeoutProvider, isTimeoutProvder := task.(TimeoutProvider); isTimeoutProvder {
				timeout := timeoutProvider.Timeout()
				if now.Sub(startedTime) >= timeout {
					jm.killHangingJob(taskName)
				}
			}
		}
	}
}

// killHangingJob cancels (sends the cancellation signal) to a running task that has exceeded its timeout.
// it assumes that the following locks are held:
// - runningTasksLock (write)
// - runningTaskStartTimesLock (write)
// - cancellationTokensLock (write)
// otherwise, chaos, mayhem, deadlocks. You should *rarely* need to call this explicitly.
func (jm *JobManager) killHangingJob(taskName string) error {
	if _, hasTask := jm.runningTasks[taskName]; hasTask {
		if token, hasToken := jm.cancellationTokens[taskName]; hasToken {
			token.Cancel()

			delete(jm.runningTasks, taskName)
			delete(jm.runningTaskStartTimes, taskName)
			delete(jm.cancellationTokens, taskName)
		}
	}
	return exception.Newf("Task name `%s` not found.", taskName)
}

// Status returns the status metadata for a JobManager
func (jm *JobManager) Status() []TaskStatus {

	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()

	jm.runningTaskStartTimesLock.RLock()
	defer jm.runningTaskStartTimesLock.RUnlock()

	jm.disabledJobsLock.RLock()
	defer jm.disabledJobsLock.RUnlock()

	jm.runningTasksLock.RLock()
	defer jm.runningTasksLock.RUnlock()

	jm.nextRunTimesLock.RLock()
	defer jm.nextRunTimesLock.RUnlock()

	jm.lastRunTimesLock.RLock()
	defer jm.lastRunTimesLock.RUnlock()

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
	jm.runningTaskStartTimesLock.RLock()
	defer jm.runningTaskStartTimesLock.RUnlock()

	jm.runningTasksLock.RLock()
	defer jm.runningTasksLock.RUnlock()

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
// Atomic Methods
// --------------------------------------------------------------------------------

func (jm *JobManager) getCancellationToken(jobName string) *CancellationToken {
	jm.cancellationTokensLock.RLock()
	defer jm.cancellationTokensLock.RUnlock()

	return jm.cancellationTokens[jobName]
}

func (jm *JobManager) setCancellationToken(jobName string, t *CancellationToken) {
	jm.cancellationTokensLock.Lock()
	defer jm.cancellationTokensLock.Unlock()

	jm.cancellationTokens[jobName] = t
}

func (jm *JobManager) deleteCancellationToken(jobName string) {
	jm.cancellationTokensLock.Lock()
	defer jm.cancellationTokensLock.Unlock()

	delete(jm.cancellationTokens, jobName)
}

func (jm *JobManager) setDisabledJob(jobName string) {
	jm.disabledJobsLock.Lock()
	defer jm.disabledJobsLock.Unlock()

	jm.disabledJobs.Add(jobName)
}

func (jm *JobManager) deleteDisabledJob(jobName string) {
	jm.disabledJobsLock.Lock()
	defer jm.disabledJobsLock.Unlock()

	jm.disabledJobs.Remove(jobName)
}

func (jm *JobManager) getNextRunTime(jobName string) *time.Time {
	jm.nextRunTimesLock.RLock()
	defer jm.nextRunTimesLock.RUnlock()

	return jm.nextRunTimes[jobName]
}

func (jm *JobManager) setNextRunTime(jobName string, t *time.Time) {
	jm.nextRunTimesLock.Lock()
	defer jm.nextRunTimesLock.Unlock()

	jm.nextRunTimes[jobName] = t
}

func (jm *JobManager) deleteNextRunTime(jobName string) {
	jm.nextRunTimesLock.Lock()
	defer jm.nextRunTimesLock.Unlock()

	delete(jm.nextRunTimes, jobName)
}

func (jm *JobManager) getLastRunTime(taskName string) time.Time {
	jm.lastRunTimesLock.RLock()
	defer jm.lastRunTimesLock.RUnlock()

	return jm.lastRunTimes[taskName]
}

func (jm *JobManager) setLastRunTime(taskName string, t time.Time) {
	jm.lastRunTimesLock.Lock()
	defer jm.lastRunTimesLock.Unlock()

	jm.lastRunTimes[taskName] = t
}

func (jm *JobManager) deleteLastRunTime(taskName string) {
	jm.lastRunTimesLock.Lock()
	defer jm.lastRunTimesLock.Unlock()

	delete(jm.lastRunTimes, taskName)
}

func (jm *JobManager) getLoadedJob(jobName string) Job {
	jm.loadedJobsLock.RLock()
	defer jm.loadedJobsLock.RUnlock()
	return jm.loadedJobs[jobName]
}

func (jm *JobManager) setLoadedJob(jobName string, j Job) {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	jm.loadedJobs[jobName] = j
}

func (jm *JobManager) deleteLoadedJob(jobName string) {
	jm.loadedJobsLock.Lock()
	defer jm.loadedJobsLock.Unlock()

	delete(jm.loadedJobs, jobName)
}

func (jm *JobManager) getRunningTask(taskName string) Task {
	jm.runningTasksLock.RLock()
	defer jm.runningTasksLock.RUnlock()

	return jm.runningTasks[taskName]
}

func (jm *JobManager) setRunningTask(taskName string, t Task) {
	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	jm.runningTasks[taskName] = t
}

func (jm *JobManager) deleteRunningTask(taskName string) {
	jm.runningTasksLock.Lock()
	defer jm.runningTasksLock.Unlock()

	delete(jm.runningTasks, taskName)
}

func (jm *JobManager) getRunningTaskStartTime(taskName string) time.Time {
	jm.runningTaskStartTimesLock.RLock()
	defer jm.runningTaskStartTimesLock.RUnlock()

	return jm.runningTaskStartTimes[taskName]
}

func (jm *JobManager) setRunningTaskStartTime(taskName string, t time.Time) {
	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	jm.runningTaskStartTimes[taskName] = t
}

func (jm *JobManager) deleteRunningTaskStartTime(taskName string) {
	jm.runningTaskStartTimesLock.Lock()
	defer jm.runningTaskStartTimesLock.Unlock()

	delete(jm.runningTaskStartTimes, taskName)
}

func (jm *JobManager) getSchedule(jobName string) Schedule {
	jm.schedulesLock.RLock()
	defer jm.schedulesLock.RUnlock()

	return jm.schedules[jobName]
}

func (jm *JobManager) setSchedule(jobName string, schedule Schedule) {
	jm.schedulesLock.Lock()
	defer jm.schedulesLock.Unlock()

	jm.schedules[jobName] = schedule
}

func (jm *JobManager) deleteSchedule(jobName string) {
	jm.schedulesLock.Lock()
	defer jm.schedulesLock.Unlock()

	delete(jm.schedules, jobName)
}
