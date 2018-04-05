package logger

import (
	"bytes"
	"encoding/json"
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

// NewErrorEvent returns a new error event.
func NewErrorEvent(flag Flag, err error) *ErrorEvent {
	return &ErrorEvent{
		flag: flag,
		ts:   time.Now().UTC(),
		err:  err,
	}
}

// NewErrorEventWithState returns a new error event with state.
func NewErrorEventWithState(flag Flag, err error, state Any) *ErrorEvent {
	return &ErrorEvent{
		flag:  flag,
		ts:    time.Now().UTC(),
		err:   err,
		state: state,
	}
}

// NewErrorEventListener returns a new error event listener.
func NewErrorEventListener(listener func(*ErrorEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*ErrorEvent); isTyped {
			listener(typed)
		}
	}
}

// ErrorEvent is an event that wraps an error.
type ErrorEvent struct {
	heading   string
	flag      Flag
	flagColor AnsiColor
	ts        time.Time
	err       error
	state     Any

	labels      map[string]string
	annotations map[string]string
}

// IsError indicates if we should write to the error writer or not.
func (e *ErrorEvent) IsError() bool {
	return true
}

// WithLabel sets a label on the event for later filtering.
func (e *ErrorEvent) WithLabel(key, value string) *ErrorEvent {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	e.labels[key] = value
	return e
}

// Labels returns a labels collection.
func (e *ErrorEvent) Labels() map[string]string {
	return e.labels
}

// WithAnnotation adds an annotation to the event.
func (e *ErrorEvent) WithAnnotation(key, value string) *ErrorEvent {
	if e.annotations == nil {
		e.annotations = map[string]string{}
	}
	e.annotations[key] = value
	return e
}

// Annotations returns the annotations set.
func (e *ErrorEvent) Annotations() map[string]string {
	return e.annotations
}

// WithTimestamp sets the event timestamp.
func (e *ErrorEvent) WithTimestamp(ts time.Time) *ErrorEvent {
	e.ts = ts
	return e
}

// Timestamp returns the event timestamp.
func (e *ErrorEvent) Timestamp() time.Time {
	return e.ts
}

// WithFlag sets the event flag.
func (e *ErrorEvent) WithFlag(flag Flag) *ErrorEvent {
	e.flag = flag
	return e
}

// Flag returns the event flag.
func (e *ErrorEvent) Flag() Flag {
	return e.flag
}

// WithHeading sets the heading.
func (e *ErrorEvent) WithHeading(heading string) *ErrorEvent {
	e.heading = heading
	return e
}

// Heading returns the heading.
func (e *ErrorEvent) Heading() string {
	return e.heading
}

// WithErr sets the error.
func (e *ErrorEvent) WithErr(err error) *ErrorEvent {
	e.err = err
	return e
}

// Err returns the underlying error.
func (e *ErrorEvent) Err() error {
	return e.err
}

// WithState sets the state.
func (e *ErrorEvent) WithState(state Any) *ErrorEvent {
	e.state = state
	return e
}

// State returns underlying state, typically an http.Request.
func (e *ErrorEvent) State() Any {
	return e.state
}

// WithFlagTextColor sets the flag text color.
func (e *ErrorEvent) WithFlagTextColor(color AnsiColor) *ErrorEvent {
	e.flagColor = color
	return e
}

// FlagTextColor returns a custom color for the flag.
func (e *ErrorEvent) FlagTextColor() AnsiColor {
	return e.flagColor
}

// WriteText implements TextWritable.
func (e *ErrorEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%+v", e.err))
}

// WriteJSON implements JSONWritable.
func (e *ErrorEvent) WriteJSON() JSONObj {
	var errorJSON Any
	if _, ok := e.err.(json.Marshaler); ok {
		errorJSON = e.err
	} else {
		errorJSON = e.err.Error()
	}
	return JSONObj{
		JSONFieldErr: errorJSON,
	}
}
