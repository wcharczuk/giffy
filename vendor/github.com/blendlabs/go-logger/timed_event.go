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
func NewTimedEventListener(listener func(e *TimedEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*TimedEvent); isTyped {
			listener(typed)
		}
	}
}

// TimedEvent is a message event with an elapsed time.
type TimedEvent struct {
	heading string
	flag    Flag
	ts      time.Time
	message string
	elapsed time.Duration

	labels      map[string]string
	annotations map[string]string
}

// WithLabel sets a label on the event for later filtering.
func (e *TimedEvent) WithLabel(key, value string) *TimedEvent {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	e.labels[key] = value
	return e
}

// Labels returns a labels collection.
func (e *TimedEvent) Labels() map[string]string {
	return e.labels
}

// WithAnnotation adds an annotation to the event.
func (e *TimedEvent) WithAnnotation(key, value string) *TimedEvent {
	if e.annotations == nil {
		e.annotations = map[string]string{}
	}
	e.annotations[key] = value
	return e
}

// Annotations returns the annotations set.
func (e *TimedEvent) Annotations() map[string]string {
	return e.annotations
}

// WithFlag sets the timed message flag.
func (e *TimedEvent) WithFlag(flag Flag) *TimedEvent {
	e.flag = flag
	return e
}

// Flag returns the timed message flag.
func (e TimedEvent) Flag() Flag {
	return e.flag
}

// WithTimestamp sets the message timestamp.
func (e *TimedEvent) WithTimestamp(ts time.Time) *TimedEvent {
	e.ts = ts
	return e
}

// Timestamp returns the timed message timestamp.
func (e TimedEvent) Timestamp() time.Time {
	return e.ts
}

// WithMessage sets the message.
func (e *TimedEvent) WithMessage(message string) *TimedEvent {
	e.message = message
	return e
}

// Message returns the string message.
func (e TimedEvent) Message() string {
	return e.message
}

// WithHeading sets the event heading.
func (e *TimedEvent) WithHeading(heading string) *TimedEvent {
	e.heading = heading
	return e
}

// Heading returns the event heading.
func (e *TimedEvent) Heading() string {
	return e.heading
}

// WithElapsed sets the elapsed time.
func (e *TimedEvent) WithElapsed(elapsed time.Duration) *TimedEvent {
	e.elapsed = elapsed
	return e
}

// Elapsed returns the elapsed time.
func (e TimedEvent) Elapsed() time.Duration {
	return e.elapsed
}

// String implements fmt.Stringer
func (e TimedEvent) String() string {
	return fmt.Sprintf("%s (%v)", e.message, e.elapsed)
}

// WriteText implements TextWritable.
func (e TimedEvent) WriteText(tf TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(e.String())
}

// WriteJSON implements JSONWritable.
func (e TimedEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldMessage: e.message,
		JSONFieldElapsed: Milliseconds(e.elapsed),
	}
}
