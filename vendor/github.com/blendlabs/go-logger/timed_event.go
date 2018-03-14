package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Timedf returns a timed message event.
func Timedf(flag Flag, elapsed time.Duration, format string, args ...Any) *TimedEvent {
	return &TimedEvent{
		flag:    flag,
		ts:      time.Now().UTC(),
		message: fmt.Sprintf(format, args...),
		elapsed: elapsed,
	}
}

// NewTimedEventListener returns a new timed event listener.
func NewTimedEventListener(listener func(te *TimedEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*TimedEvent); isTyped {
			listener(typed)
		}
	}
}

// TimedEvent is a message event with an elapsed time.
type TimedEvent struct {
	flag    Flag
	ts      time.Time
	label   string
	message string
	elapsed time.Duration
}

// WithFlag sets the timed message flag.
func (te *TimedEvent) WithFlag(flag Flag) *TimedEvent {
	te.flag = flag
	return te
}

// Flag returns the timed message flag.
func (te TimedEvent) Flag() Flag {
	return te.flag
}

// WithTimestamp sets the message timestamp.
func (te *TimedEvent) WithTimestamp(ts time.Time) *TimedEvent {
	te.ts = ts
	return te
}

// Timestamp returns the timed message timestamp.
func (te TimedEvent) Timestamp() time.Time {
	return te.ts
}

// WithLabel sets the label.
func (te *TimedEvent) WithLabel(label string) *TimedEvent {
	te.label = label
	return te
}

// Label returns the label.
func (te TimedEvent) Label() string {
	return te.label
}

// WithMessage sets the message.
func (te *TimedEvent) WithMessage(message string) *TimedEvent {
	te.message = message
	return te
}

// Message returns the string message.
func (te TimedEvent) Message() string {
	return te.message
}

// WithElapsed sets the elapsed time.
func (te *TimedEvent) WithElapsed(elapsed time.Duration) *TimedEvent {
	te.elapsed = elapsed
	return te
}

// Elapsed returns the elapsed time.
func (te TimedEvent) Elapsed() time.Duration {
	return te.elapsed
}

// String implements fmt.Stringer
func (te TimedEvent) String() string {
	return fmt.Sprintf("%s (%v)", te.message, te.elapsed)
}

// WriteText implements TextWritable.
func (te TimedEvent) WriteText(tf TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(te.String())
}

// WriteJSON implements JSONWritable.
func (te TimedEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldMessage: te.message,
		JSONFieldElapsed: Milliseconds(te.elapsed),
	}
}
