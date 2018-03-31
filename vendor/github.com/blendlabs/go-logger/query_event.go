package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

const (
	// Query is a logging flag.
	Query Flag = "db.query"
)

// NewQueryEvent creates a new query event.
func NewQueryEvent(body string, elapsed time.Duration) *QueryEvent {
	return &QueryEvent{
		flag:    Query,
		ts:      time.Now().UTC(),
		body:    body,
		elapsed: elapsed,
	}
}

// NewQueryEventListener returns a new listener for spiffy events.
func NewQueryEventListener(listener func(e *QueryEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*QueryEvent); isTyped {
			listener(typed)
		}
	}
}

// QueryEvent represents a database query.
type QueryEvent struct {
	flag     Flag
	ts       time.Time
	engine   string
	label    string
	body     string
	database string
	elapsed  time.Duration
}

// WithFlag sets the flag.
func (e *QueryEvent) WithFlag(flag Flag) *QueryEvent {
	e.flag = flag
	return e
}

// Flag returns the event flag.
func (e QueryEvent) Flag() Flag {
	return e.flag
}

// WithTimestamp sets the timestamp.
func (e *QueryEvent) WithTimestamp(ts time.Time) *QueryEvent {
	e.ts = ts
	return e
}

// Timestamp returns the event timestamp.
func (e QueryEvent) Timestamp() time.Time {
	return e.ts
}

// WithEngine sets the engine.
func (e *QueryEvent) WithEngine(engine string) *QueryEvent {
	e.engine = engine
	return e
}

// Engine returns the engine.
func (e QueryEvent) Engine() string {
	return e.engine
}

// WithDatabase sets the database.
func (e *QueryEvent) WithDatabase(db string) *QueryEvent {
	e.database = db
	return e
}

// Database returns the event database.
func (e QueryEvent) Database() string {
	return e.database
}

// WithLabel sets the label.
func (e *QueryEvent) WithLabel(label string) *QueryEvent {
	e.label = label
	return e
}

// Label returns the query label.
func (e QueryEvent) Label() string {
	return e.label
}

// WithBody sets the body.
func (e *QueryEvent) WithBody(body string) *QueryEvent {
	e.body = body
	return e
}

// Body returns the query body.
func (e QueryEvent) Body() string {
	return e.body
}

// WithElapsed sets the elapsed time.
func (e *QueryEvent) WithElapsed(elapsed time.Duration) *QueryEvent {
	e.elapsed = elapsed
	return e
}

// Elapsed returns the elapsed time.
func (e QueryEvent) Elapsed() time.Duration {
	return e.elapsed
}

// WriteText writes the event text to the output.
func (e QueryEvent) WriteText(tf TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("[%s] (%v)", tf.Colorize(e.database, ColorBlue), e.elapsed))
	if len(e.label) > 0 {
		buf.WriteRune(RuneSpace)
		buf.WriteString(e.label)
	}
	if len(e.body) > 0 {
		buf.WriteRune(RuneSpace)
		buf.WriteString(strings.TrimSpace(e.body))
	}
}

// WriteJSON implements JSONWritable.
func (e QueryEvent) WriteJSON() JSONObj {
	return JSONObj{
		"engine":         e.engine,
		"database":       e.database,
		"label":          e.label,
		"body":           e.body,
		JSONFieldElapsed: Milliseconds(e.elapsed),
	}
}
