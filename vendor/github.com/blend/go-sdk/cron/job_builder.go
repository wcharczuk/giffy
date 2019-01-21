package cron

import (
	"context"
	"time"
)

// Interface assertions.
var (
	_ ScheduleProvider               = (*JobBuilder)(nil)
	_ TimeoutProvider                = (*JobBuilder)(nil)
	_ EnabledProvider                = (*JobBuilder)(nil)
	_ ShouldWriteOutputProvider      = (*JobBuilder)(nil)
	_ ShouldTriggerListenersProvider = (*JobBuilder)(nil)
	_ OnStartReceiver                = (*JobBuilder)(nil)
	_ OnCancellationReceiver         = (*JobBuilder)(nil)
	_ OnCompleteReceiver             = (*JobBuilder)(nil)
	_ OnFailureReceiver              = (*JobBuilder)(nil)
	_ OnBrokenReceiver               = (*JobBuilder)(nil)
	_ OnFixedReceiver                = (*JobBuilder)(nil)
)

// NewJob returns a new job factory.
func NewJob(name string) *JobBuilder {
	return &JobBuilder{
		name: name,
	}
}

// JobBuilder allows for job creation w/o a fully formed struct.
type JobBuilder struct {
	name                           string
	timeoutProvider                func() time.Duration
	enabledProvider                func() bool
	shouldTriggerListenersProvider func() bool
	shouldWriteOutputProvider      func() bool
	schedule                       Schedule
	action                         Action

	onStart        func(*JobInvocation)
	onCancellation func(*JobInvocation)
	onComplete     func(*JobInvocation)
	onFailure      func(*JobInvocation)
	onBroken       func(*JobInvocation)
	onFixed        func(*JobInvocation)
}

// WithName sets the job name.
func (jb *JobBuilder) WithName(name string) *JobBuilder {
	jb.name = name
	return jb
}

// WithSchedule sets the schedule for the job.
func (jb *JobBuilder) WithSchedule(schedule Schedule) *JobBuilder {
	jb.schedule = schedule
	return jb
}

// WithTimeoutProvider sets the timeout provider.
func (jb *JobBuilder) WithTimeoutProvider(timeoutProvider func() time.Duration) *JobBuilder {
	jb.timeoutProvider = timeoutProvider
	return jb
}

// WithAction sets the job action.
func (jb *JobBuilder) WithAction(action Action) *JobBuilder {
	jb.action = action
	return jb
}

// WithEnabledProvider sets the enabled provider for the job.
func (jb *JobBuilder) WithEnabledProvider(enabledProvider func() bool) *JobBuilder {
	jb.enabledProvider = enabledProvider
	return jb
}

// WithShouldTriggerListenersProvider sets the enabled provider for the job.
func (jb *JobBuilder) WithShouldTriggerListenersProvider(provider func() bool) *JobBuilder {
	jb.shouldTriggerListenersProvider = provider
	return jb
}

// WithShouldWriteOutputProvider sets the enabled provider for the job.
func (jb *JobBuilder) WithShouldWriteOutputProvider(provider func() bool) *JobBuilder {
	jb.shouldWriteOutputProvider = provider
	return jb
}

// WithOnStart sets a lifecycle handler.
func (jb *JobBuilder) WithOnStart(receiver func(*JobInvocation)) *JobBuilder {
	jb.onStart = receiver
	return jb
}

// WithOnCancellation sets a lifecycle handler.
func (jb *JobBuilder) WithOnCancellation(receiver func(*JobInvocation)) *JobBuilder {
	jb.onCancellation = receiver
	return jb
}

// WithOnComplete sets a lifecycle handler.
func (jb *JobBuilder) WithOnComplete(receiver func(*JobInvocation)) *JobBuilder {
	jb.onComplete = receiver
	return jb
}

// WithOnFailure sets a lifecycle handler.
func (jb *JobBuilder) WithOnFailure(receiver func(*JobInvocation)) *JobBuilder {
	jb.onFailure = receiver
	return jb
}

// WithOnFixed sets a lifecycle handler.
func (jb *JobBuilder) WithOnFixed(receiver func(*JobInvocation)) *JobBuilder {
	jb.onFixed = receiver
	return jb
}

// WithOnBroken sets a lifecycle handler.
func (jb *JobBuilder) WithOnBroken(receiver func(*JobInvocation)) *JobBuilder {
	jb.onBroken = receiver
	return jb
}

//
// implementations of interface methods
//

// Name returns the job name.
func (jb *JobBuilder) Name() string {
	return jb.name
}

// Schedule returns the job schedule.
func (jb *JobBuilder) Schedule() Schedule {
	return jb.schedule
}

// Timeout returns the job timeout.
func (jb *JobBuilder) Timeout() (timeout time.Duration) {
	if jb.timeoutProvider != nil {
		return jb.timeoutProvider()
	}
	return
}

// Enabled returns if the job is enabled.
func (jb *JobBuilder) Enabled() bool {
	if jb.enabledProvider != nil {
		return jb.enabledProvider()
	}
	return true
}

// ShouldWriteOutput implements the should write output provider.
func (jb *JobBuilder) ShouldWriteOutput() bool {
	if jb.shouldWriteOutputProvider != nil {
		return jb.shouldWriteOutputProvider()
	}
	return true
}

// ShouldTriggerListeners implements the should trigger listeners provider.
func (jb *JobBuilder) ShouldTriggerListeners() bool {
	if jb.shouldTriggerListenersProvider != nil {
		return jb.shouldTriggerListenersProvider()
	}
	return true
}

// OnStart is a lifecycle hook.
func (jb *JobBuilder) OnStart(ctx context.Context) {
	if jb.onStart != nil {
		jb.onStart(GetJobInvocation(ctx))
	}
}

// OnCancellation is a lifecycle hook.
func (jb *JobBuilder) OnCancellation(ctx context.Context) {
	if jb.onCancellation != nil {
		jb.onCancellation(GetJobInvocation(ctx))
	}
}

// OnComplete is a lifecycle hook.
func (jb *JobBuilder) OnComplete(ctx context.Context) {
	if jb.onComplete != nil {
		jb.onComplete(GetJobInvocation(ctx))
	}
}

// OnFailure is a lifecycle hook.
func (jb *JobBuilder) OnFailure(ctx context.Context) {
	if jb.onFailure != nil {
		jb.onFailure(GetJobInvocation(ctx))
	}
}

// OnFixed is a lifecycle hook.
func (jb *JobBuilder) OnFixed(ctx context.Context) {
	if jb.onFixed != nil {
		jb.onFixed(GetJobInvocation(ctx))
	}
}

// OnBroken is a lifecycle hook.
func (jb *JobBuilder) OnBroken(ctx context.Context) {
	if jb.onBroken != nil {
		jb.onBroken(GetJobInvocation(ctx))
	}
}

// Execute runs the job action if it's set.
func (jb *JobBuilder) Execute(ctx context.Context) error {
	if jb.action != nil {
		return jb.action(ctx)
	}
	return nil
}
