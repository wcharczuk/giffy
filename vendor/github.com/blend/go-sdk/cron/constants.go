package cron

import (
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

const (
	// DefaultHeartbeatInterval is the interval between schedule next run checks.
	DefaultHeartbeatInterval = 50 * time.Millisecond

	// DefaultHighPrecisionHeartbeatInterval is the high precision interval between schedule next run checks.
	DefaultHighPrecisionHeartbeatInterval = 5 * time.Millisecond
)

const (
	// FlagStarted is an event flag.
	FlagStarted logger.Flag = "cron.started"
	// FlagComplete is an event flag.
	FlagComplete logger.Flag = "cron.complete"
)

const (
	// EnvVarHeartbeatInterval is an environment variable name.
	EnvVarHeartbeatInterval = "CRON_HEARTBEAT_INTERVAL"
)

const (
	// ErrJobNotLoaded is a common error.
	ErrJobNotLoaded Error = "job not loaded"

	// ErrJobAlreadyLoaded is a common error.
	ErrJobAlreadyLoaded Error = "job already loaded"

	// ErrTaskNotFound is a common error.
	ErrTaskNotFound Error = "task not found"
)

// IsJobNotLoaded returns if the error is a job not loaded error.
func IsJobNotLoaded(err error) bool {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err == ErrJobNotLoaded
}

// IsJobAlreadyLoaded returns if the error is a job already loaded error.
func IsJobAlreadyLoaded(err error) bool {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err == ErrJobAlreadyLoaded
}

// IsTaskNotFound returns if the error is a task not found error.
func IsTaskNotFound(err error) bool {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err == ErrTaskNotFound
}

// State is a job state.
type State string

const (
	//StateRunning is the running state.
	StateRunning State = "running"

	// StateEnabled is the enabled state.
	StateEnabled State = "enabled"

	// StateDisabled is the disabled state.
	StateDisabled State = "disabled"
)
