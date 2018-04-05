package cron

import (
	"bytes"
	"fmt"
	"time"

	logger "github.com/blend/go-sdk/logger"
)

const (
	// FlagStarted is a logger flag for task start.
	FlagStarted logger.Flag = "chronometer.task"
	// FlagComplete is a logger flag for task completions.
	FlagComplete logger.Flag = "chronometer.task.complete"
)

// NewEventStartedListener returns a new event started listener.
func NewEventStartedListener(listener func(e EventStarted)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(EventStarted); isTyped {
			listener(typed)
		}
	}
}

// EventStarted is a started event.
type EventStarted struct {
	ts         time.Time
	isEnabled  bool
	isWritable bool
	taskName   string
}

// Flag returns the event flag.
func (e EventStarted) Flag() logger.Flag {
	return FlagStarted
}

// Timestamp returns an event timestamp.
func (e EventStarted) Timestamp() time.Time {
	return e.ts
}

// IsEnabled determines if the event triggers listeners.
func (e EventStarted) IsEnabled() bool {
	return e.isEnabled
}

// IsWritable determines if the event is written to the logger output.
func (e EventStarted) IsWritable() bool {
	return e.isWritable
}

// TaskName returns the event task name.
func (e EventStarted) TaskName() string {
	return e.taskName
}

// WriteText implements logger.TextWritable.
func (e EventStarted) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("`%s` starting", e.taskName))
}

// WriteJSON implements logger.JSONWritable.
func (e EventStarted) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"taskName": e.taskName,
	}
}

// NewEventCompleteListener returns a new event complete listener.
func NewEventCompleteListener(listener func(e EventComplete)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(EventComplete); isTyped {
			listener(typed)
		}
	}
}

// EventComplete is an event emitted to the logger.
type EventComplete struct {
	ts         time.Time
	isEnabled  bool
	isWritable bool
	taskName   string
	err        error
	elapsed    time.Duration
}

// Flag returns the event flag.
func (e EventComplete) Flag() logger.Flag {
	return FlagComplete
}

// Timestamp returns an event timestamp.
func (e EventComplete) Timestamp() time.Time {
	return e.ts
}

// IsEnabled determines if the event triggers listeners.
func (e EventComplete) IsEnabled() bool {
	return e.isEnabled
}

// IsWritable determines if the event is written to the logger output.
func (e EventComplete) IsWritable() bool {
	return e.isWritable
}

// TaskName returns the event task name.
func (e EventComplete) TaskName() string {
	return e.taskName
}

// Elapsed returns the elapsed time for the task.
func (e EventComplete) Elapsed() time.Duration {
	return e.elapsed
}

// Err returns the event err (if any).
func (e EventComplete) Err() error {
	return e.err
}

// WriteText implements logger.TextWritable.
func (e EventComplete) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	if e.err != nil {
		buf.WriteString(fmt.Sprintf("`%s` failed (%v)", e.taskName, e.elapsed))
	} else {
		buf.WriteString(fmt.Sprintf("`%s` completed (%v)", e.taskName, e.elapsed))
	}
}

// WriteJSON implements logger.JSONWritable.
func (e EventComplete) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"taskName":              e.taskName,
		logger.JSONFieldElapsed: logger.Milliseconds(e.elapsed),
		logger.JSONFieldErr:     e.err,
	}
}
