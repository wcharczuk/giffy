package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Messagef returns a new Message Event.
func Messagef(flag Flag, format string, args ...Any) *MessageEvent {
	return &MessageEvent{
		flag:    flag,
		ts:      time.Now().UTC(),
		message: fmt.Sprintf(format, args...),
	}
}

// MessagefWithFlagTextColor returns a new Message Event with a given flag text color.
func MessagefWithFlagTextColor(flag Flag, flagColor AnsiColor, format string, args ...Any) *MessageEvent {
	return &MessageEvent{
		flag:      flag,
		flagColor: flagColor,
		ts:        time.Now().UTC(),
		message:   fmt.Sprintf(format, args...),
	}
}

// NewMessageEventListener returns a new message event listener.
func NewMessageEventListener(listener func(*MessageEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*MessageEvent); isTyped {
			listener(typed)
		}
	}
}

// MessageEvent is a common type of message.
type MessageEvent struct {
	heading   string
	flag      Flag
	flagColor AnsiColor
	ts        time.Time
	message   string

	labels      map[string]string
	annotations map[string]string
}

// WithLabel sets a label on the event for later filtering.
func (e *MessageEvent) WithLabel(key, value string) *MessageEvent {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	e.labels[key] = value
	return e
}

// Labels returns a labels collection.
func (e *MessageEvent) Labels() map[string]string {
	return e.labels
}

// WithAnnotation adds an annotation to the event.
func (e *MessageEvent) WithAnnotation(key, value string) *MessageEvent {
	if e.annotations == nil {
		e.annotations = map[string]string{}
	}
	e.annotations[key] = value
	return e
}

// Annotations returns the annotations set.
func (e *MessageEvent) Annotations() map[string]string {
	return e.annotations
}

// WithFlag sets the message flag.
func (e *MessageEvent) WithFlag(flag Flag) *MessageEvent {
	e.flag = flag
	return e
}

// Flag returns the message flag.
func (e *MessageEvent) Flag() Flag {
	return e.flag
}

// WithTimestamp sets the message timestamp.
func (e *MessageEvent) WithTimestamp(ts time.Time) *MessageEvent {
	e.ts = ts
	return e
}

// Timestamp returns the message timestamp.
func (e *MessageEvent) Timestamp() time.Time {
	return e.ts
}

// WithMessage sets the message.
func (e *MessageEvent) WithMessage(message string) *MessageEvent {
	e.message = message
	return e
}

// Message returns the message.
func (e *MessageEvent) Message() string {
	return e.message
}

// WithHeading sets the heading.
func (e *MessageEvent) WithHeading(heading string) *MessageEvent {
	e.heading = heading
	return e
}

// Heading returns the heading.
func (e *MessageEvent) Heading() string {
	return e.heading
}

// WithFlagTextColor sets the message flag text color.
func (e *MessageEvent) WithFlagTextColor(color AnsiColor) *MessageEvent {
	e.flagColor = color
	return e
}

// FlagTextColor returns a custom color for the flag.
func (e *MessageEvent) FlagTextColor() AnsiColor {
	return e.flagColor
}

// WriteText implements TextWritable.
func (e *MessageEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(e.message)
}

// WriteJSON implements JSONWriteable.
func (e *MessageEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldMessage: e.message,
	}
}

// String returns the message event body.
func (e *MessageEvent) String() string {
	return e.message
}
