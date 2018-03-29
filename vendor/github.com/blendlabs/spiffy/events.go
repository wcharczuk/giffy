package spiffy

import (
	"bytes"
	"fmt"
	"strings"
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
func NewEvent(flag logger.Flag, label, database, queryBody string, elapsed time.Duration, err error) Event {
	return Event{
		flag:       flag,
		ts:         time.Now().UTC(),
		queryLabel: label,
		database:   database,
		queryBody:  queryBody,
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
	database   string
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
	buf.WriteString(fmt.Sprintf("[%s] (%v)", tf.Colorize(e.database, logger.ColorBlue), e.elapsed))
	if len(e.queryLabel) > 0 {
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(e.queryLabel)
	}
	if len(e.queryBody) > 0 {
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(strings.TrimSpace(e.queryBody))
	}
}

// WriteJSON implements logger.JSONWritable.
func (e Event) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"database":              e.database,
		"queryLabel":            e.queryLabel,
		"queryBody":             e.queryBody,
		logger.JSONFieldElapsed: logger.Milliseconds(e.elapsed),
	}
}
