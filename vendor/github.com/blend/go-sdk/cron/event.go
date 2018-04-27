package cron

import (
	"bytes"
	"fmt"
	"time"

	logger "github.com/blend/go-sdk/logger"
)

// these are compile time assertions
var (
	_ logger.Event            = &Event{}
	_ logger.EventHeadings    = &Event{}
	_ logger.EventLabels      = &Event{}
	_ logger.EventAnnotations = &Event{}
)

// NewEventListener returns a new event listener.
func NewEventListener(listener func(e *Event)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(*Event); isTyped {
			listener(typed)
		}
	}
}

// NewEvent creates a new event.
func NewEvent(flag logger.Flag, taskName string) *Event {
	return &Event{
		EventMeta: logger.NewEventMeta(flag),
		taskName:  taskName,
		enabled:   true,
		writable:  true,
	}
}

// Event is an event.
type Event struct {
	*logger.EventMeta

	complete bool

	enabled  bool
	writable bool

	taskName string
	err      error
	elapsed  time.Duration
}

// WithHeadings sets the headings.
func (e *Event) WithHeadings(headings ...string) *Event {
	e.SetHeadings(headings...)
	return e
}

// WithLabel sets a label on the event for later filtering.
func (e *Event) WithLabel(key, value string) *Event {
	e.AddLabelValue(key, value)
	return e
}

// WithAnnotation adds an annotation to the event.
func (e *Event) WithAnnotation(key, value string) *Event {
	e.AddAnnotationValue(key, value)
	return e
}

// WithFlag sets the event flag.
func (e *Event) WithFlag(f logger.Flag) *Event {
	e.SetFlag(f)
	return e
}

// WithTimestamp sets the message timestamp.
func (e *Event) WithTimestamp(ts time.Time) *Event {
	e.SetTimestamp(ts)
	return e
}

// WithIsEnabled sets if the event is enabled
func (e *Event) WithIsEnabled(isEnabled bool) *Event {
	e.enabled = isEnabled
	return e
}

// IsEnabled determines if the event triggers listeners.
func (e Event) IsEnabled() bool {
	return e.enabled
}

// WithIsWritable sets if the event is writable.
func (e *Event) WithIsWritable(isWritable bool) *Event {
	e.writable = isWritable
	return e
}

// IsWritable determines if the event is written to the logger output.
func (e Event) IsWritable() bool {
	return e.writable
}

// WithTaskName sets the task name.
func (e *Event) WithTaskName(taskName string) *Event {
	e.taskName = taskName
	return e
}

// TaskName returns the event task name.
func (e Event) TaskName() string {
	return e.taskName
}

// WithErr sets the error on the event.
func (e *Event) WithErr(err error) *Event {
	e.err = err
	return e
}

// Err returns the event err (if any).
func (e Event) Err() error {
	return e.err
}

// Complete returns if the event completed.
func (e Event) Complete() bool {
	return e.Flag() == FlagComplete
}

// WithElapsed sets the elapsed time.
func (e *Event) WithElapsed(d time.Duration) *Event {
	e.elapsed = d
	return e
}

// Elapsed returns the elapsed time for the task.
func (e Event) Elapsed() time.Duration {
	return e.elapsed
}

// WriteText implements logger.TextWritable.
func (e Event) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	if e.Complete() {
		if e.err != nil {
			buf.WriteString(fmt.Sprintf("`%s` failed (%v)", e.taskName, e.elapsed))
		} else {
			buf.WriteString(fmt.Sprintf("`%s` completed (%v)", e.taskName, e.elapsed))
		}
	} else {
		buf.WriteString(fmt.Sprintf("`%s` started", e.taskName))
	}
}

// WriteJSON implements logger.JSONWritable.
func (e Event) WriteJSON() logger.JSONObj {
	obj := logger.JSONObj{
		"taskName": e.taskName,
	}
	if e.err != nil {
		obj[logger.JSONFieldErr] = e.err
	}
	if e.elapsed > 0 {
		obj[logger.JSONFieldElapsed] = logger.Milliseconds(e.elapsed)
	}
	return obj
}
