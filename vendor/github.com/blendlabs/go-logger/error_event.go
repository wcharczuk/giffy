package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Errorf returns a new error event based on format and arguments.
func Errorf(flag Flag, format string, args ...Any) *ErrorEvent {
	return &ErrorEvent{
		flag: flag,
		ts:   time.Now().UTC(),
		err:  fmt.Errorf(format, args...),
	}
}

// ErrorfWithFlagTextColor returns a new error event based on format and arguments with a given flag text color.
func ErrorfWithFlagTextColor(flag Flag, flagColor AnsiColor, format string, args ...Any) *ErrorEvent {
	return &ErrorEvent{
		flag:      flag,
		flagColor: flagColor,
		ts:        time.Now().UTC(),
		err:       fmt.Errorf(format, args...),
	}
}

// NewError returns a new error event.
func NewError(flag Flag, err error) *ErrorEvent {
	return &ErrorEvent{
		flag: flag,
		ts:   time.Now().UTC(),
		err:  err,
	}
}

// NewErrorWithState returns a new error event with state.
func NewErrorWithState(flag Flag, err error, state Any) *ErrorEvent {
	return &ErrorEvent{
		flag:  flag,
		ts:    time.Now().UTC(),
		err:   err,
		state: state,
	}
}

// NewErrorEventListener returns a new error event listener.
func NewErrorEventListener(listener func(me *ErrorEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*ErrorEvent); isTyped {
			listener(typed)
		}
	}
}

// ErrorEvent is an event that wraps an error.
type ErrorEvent struct {
	flag      Flag
	flagColor AnsiColor
	ts        time.Time
	label     string
	err       error
	state     Any
}

// IsError indicates if we should write to the error writer or not.
func (ee ErrorEvent) IsError() bool {
	return true
}

// WithTimestamp sets the event timestamp.
func (ee *ErrorEvent) WithTimestamp(ts time.Time) *ErrorEvent {
	ee.ts = ts
	return ee
}

// Timestamp returns the event timestamp.
func (ee ErrorEvent) Timestamp() time.Time {
	return ee.ts
}

// WithFlag sets the event flag.
func (ee *ErrorEvent) WithFlag(flag Flag) *ErrorEvent {
	ee.flag = flag
	return ee
}

// Flag returns the event flag.
func (ee ErrorEvent) Flag() Flag {
	return ee.flag
}

// WithLabel sets the label.
func (ee *ErrorEvent) WithLabel(label string) *ErrorEvent {
	ee.label = label
	return ee
}

// Label returns the label.
func (ee ErrorEvent) Label() string {
	return ee.label
}

// WithErr sets the error.
func (ee *ErrorEvent) WithErr(err error) *ErrorEvent {
	ee.err = err
	return ee
}

// Err returns the underlying error.
func (ee ErrorEvent) Err() error {
	return ee.err
}

// WithState sets the state.
func (ee *ErrorEvent) WithState(state Any) *ErrorEvent {
	ee.state = state
	return ee
}

// State returns underlying state, typically an http.Request.
func (ee ErrorEvent) State() Any {
	return ee.state
}

// WithFlagTextColor sets the flag text color.
func (ee *ErrorEvent) WithFlagTextColor(color AnsiColor) *ErrorEvent {
	ee.flagColor = color
	return ee
}

// FlagTextColor returns a custom color for the flag.
func (ee ErrorEvent) FlagTextColor() AnsiColor {
	return ee.flagColor
}

// WriteText implements TextWritable.
func (ee ErrorEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%+v", ee.err))
}

// WriteJSON implements JSONWritable.
func (ee ErrorEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldErr: ee.err,
	}
}
