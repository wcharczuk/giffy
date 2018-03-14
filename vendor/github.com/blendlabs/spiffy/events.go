package spiffy

import (
	"bytes"
	"fmt"
	"time"

	logger "github.com/blendlabs/go-logger"
)

const (
	// FlagExecute is a logger.EventFlag
	FlagExecute logger.Flag = "db.execute"

	// FlagQuery is a logger.EventFlag
	FlagQuery logger.Flag = "db.query"
)

// NewEvent creates a new logger event.
func NewEvent(flag logger.Flag, label string, elapsed time.Duration, err error) Event {
	return Event{
		flag:       flag,
		ts:         time.Now().UTC(),
		queryLabel: label,
		elapsed:    elapsed,
		err:        err,
	}
}

// NewEventListener returns a new listener for spiffy events.
func NewEventListener(listener func(e Event)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(Event); isTyped {
			listener(typed)
		}
	}
}

// Event is the event we trigger the logger with.
type Event struct {
	flag       logger.Flag
	ts         time.Time
	queryLabel string
	queryBody  string
	elapsed    time.Duration
	err        error
}

// Flag returns the event flag.
func (e Event) Flag() logger.Flag {
	return e.flag
}

// Timestamp returns the event timestamp.
func (e Event) Timestamp() time.Time {
	return e.ts
}

// QueryLabel returns the query label.
func (e Event) QueryLabel() string {
	return e.queryLabel
}

// Elapsed returns the elapsed time.
func (e Event) Elapsed() time.Duration {
	return e.elapsed
}

// Err returns the error.
func (e Event) Err() error {
	return e.err
}

// WriteText writes the event text to the output.
func (e Event) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("(%v) ", e.elapsed))
	if len(e.queryLabel) > 0 {
		buf.WriteString(e.queryLabel)
	}
	buf.WriteRune(logger.RuneNewline)
}

// WriteJSON implements logger.JSONWritable.
func (e Event) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"queryLabel":            e.queryLabel,
		logger.JSONFieldElapsed: logger.Milliseconds(e.elapsed),
	}
}

// NewStatementEvent creates a new logger event.
func NewStatementEvent(flag logger.Flag, label, query string, elapsed time.Duration, err error) StatementEvent {
	return StatementEvent{
		Event:     NewEvent(flag, label, elapsed, err),
		queryBody: query,
	}
}

// NewStatementEventListener returns a new listener for spiffy statement events.
func NewStatementEventListener(listener func(e StatementEvent)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(StatementEvent); isTyped {
			listener(typed)
		}
	}
}

// StatementEvent is the event we trigger the logger with.
type StatementEvent struct {
	Event
	queryBody string
}

// QueryBody returns the query body.
func (e StatementEvent) QueryBody() string {
	return e.queryBody
}

// WriteText writes the event text to the output.
func (e StatementEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("(%v) ", e.elapsed))
	if len(e.queryLabel) > 0 {
		buf.WriteString(e.queryLabel)
	}
	buf.WriteRune(logger.RuneNewline)
	buf.WriteString(e.queryBody)
}

// WriteJSON implements logger.JSONWritable.
func (e StatementEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"queryLabel":            e.queryLabel,
		"queryBody":             e.queryBody,
		logger.JSONFieldElapsed: logger.Milliseconds(e.elapsed),
	}
}
