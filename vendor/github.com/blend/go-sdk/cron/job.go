package cron

import (
	"context"
	"time"
)

// Job is an interface structs can satisfy to be loaded into the JobManager.
type Job interface {
	Name() string
	Schedule() Schedule
	Execute(ctx context.Context) error
}

// IsJobCancelled check if a job is cancelled
func IsJobCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// EventTriggerListenersProvider is a type that enables or disables logger listeners.
type EventTriggerListenersProvider interface {
	ShouldTriggerListeners() bool
}

// EventShouldWriteOutputProvider is a type that enables or disables logger output for events.
type EventShouldWriteOutputProvider interface {
	ShouldWriteOutput() bool
}

// EnabledProvider is an optional interface that will allow jobs to control if they're enabled.
type EnabledProvider interface {
	Enabled() bool
}

// NewJob returns a new job factory.
func NewJob() *JobFactory {
	return &JobFactory{
		schedule: OnDemand(),
	}
}

// JobFactory allows for job creation w/o a fully formed struct.
type JobFactory struct {
	name                 string
	schedule             Schedule
	timeout              time.Duration
	action               TaskAction
	isEnabledProvider    func() bool
	showMessagesProvider func() bool
}

// Name returns the job name.
func (jf *JobFactory) Name() string {
	return jf.name
}

// WithName sets the job name.
func (jf *JobFactory) WithName(name string) *JobFactory {
	jf.name = name
	return jf
}

// Schedule returns the job schedule.
func (jf *JobFactory) Schedule() Schedule {
	return jf.schedule
}

// WithSchedule sets the schedule for the job.
func (jf *JobFactory) WithSchedule(schedule Schedule) *JobFactory {
	jf.schedule = schedule
	return jf
}

// Timeout returns the job timeout.
func (jf *JobFactory) Timeout() time.Duration {
	return jf.timeout
}

// WithTimeout sets the timeout.
func (jf *JobFactory) WithTimeout(timeout time.Duration) *JobFactory {
	jf.timeout = timeout
	return jf
}

// Execute runs the job action if it's set.
func (jf *JobFactory) Execute(ctx context.Context) error {
	if jf.action != nil {
		return jf.action(ctx)
	}
	return nil
}

// WithAction sets the job action.
func (jf *JobFactory) WithAction(action TaskAction) *JobFactory {
	jf.action = action
	return jf
}

// Action returns the job action.
func (jf *JobFactory) Action() TaskAction {
	return jf.action
}

// WithIsEnabledProvider sets the enabled provider for the job.
func (jf *JobFactory) WithIsEnabledProvider(provider func() bool) *JobFactory {
	jf.isEnabledProvider = provider
	return jf
}

// IsEnabled returns if the job is enabled.
func (jf *JobFactory) IsEnabled() bool {
	if jf.isEnabledProvider != nil {
		return jf.isEnabledProvider()
	}
	return true
}

// WithShowMessagesProvider sets the enabled provider for the job.
func (jf *JobFactory) WithShowMessagesProvider(provider func() bool) *JobFactory {
	jf.showMessagesProvider = provider
	return jf
}

// ShowMessages returns if the job should trigger logging events.
func (jf *JobFactory) ShowMessages() bool {
	if jf.showMessagesProvider != nil {
		return jf.showMessagesProvider()
	}
	return true
}
