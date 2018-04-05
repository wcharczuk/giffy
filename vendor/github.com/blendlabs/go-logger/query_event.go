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
	heading string

	flag       Flag
	ts         time.Time
	engine     string
	queryLabel string
	body       string
	database   string
	elapsed    time.Duration

	labels      map[string]string
	annotations map[string]string
}

// WithHeading sets the event heading.
func (e *QueryEvent) WithHeading(heading string) *QueryEvent {
	e.heading = heading
	return e
}

// Heading returns the event heading.
func (e *QueryEvent) Heading() string {
	return e.heading
}

// WithLabel sets a label on the event for later filtering.
func (e *QueryEvent) WithLabel(key, value string) *QueryEvent {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	e.labels[key] = value
	return e
}

// Labels returns a labels collection.
func (e *QueryEvent) Labels() map[string]string {
	return e.labels
}

// WithAnnotation adds an annotation to the event.
func (e *QueryEvent) WithAnnotation(key, value string) *QueryEvent {
	if e.annotations == nil {
		e.annotations = map[string]string{}
	}
	e.annotations[key] = value
	return e
}

// Annotations returns the annotations set.
func (e *QueryEvent) Annotations() map[string]string {
	return e.annotations
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

// WithQueryLabel sets the query label.
func (e *QueryEvent) WithQueryLabel(queryLabel string) *QueryEvent {
	e.queryLabel = queryLabel
	return e
}

// QueryLabel returns the query label.
func (e QueryEvent) QueryLabel() string {
	return e.queryLabel
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
	if len(e.queryLabel) > 0 {
		buf.WriteRune(RuneSpace)
		buf.WriteString(e.queryLabel)
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
		"queryLabel":     e.queryLabel,
		"body":           e.body,
		JSONFieldElapsed: Milliseconds(e.elapsed),
	}
}
