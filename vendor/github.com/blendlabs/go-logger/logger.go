package logger

import (
	"bytes"
	"io"
)

type loggerOutputWithTimeSource func(ts TimeSource, format string, args ...interface{})

// Logger is the basic interface to a logger implementation.
type Logger interface {
	Printf(format string, args ...interface{})
	PrintfWithTimeSource(ts TimeSource, format string, args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorfWithTimeSource(ts TimeSource, format string, args ...interface{})
	Fprintf(w io.Writer, format string, args ...interface{})
	FprintfWithTimeSource(ts TimeSource, w io.Writer, format string, args ...interface{})
	Write(data []byte) (int64, error)
	WriteWithTimeSource(ts TimeSource, data []byte) (int64, error)

	Colorize(value string, color AnsiColorCode) string
	ColorizeByStatusCode(statusCode int, value string) string

	GetBuffer() *bytes.Buffer
	PutBuffer(*bytes.Buffer)

	UseAnsiColors() bool
	SetUseAnsiColors(useAnsiColors bool)

	ShowTimestamp() bool
	SetShowTimestamp(showTimestamp bool)

	ShowLabel() bool
	SetShowLabel(showLabel bool)

	Label() string
	SetLabel(label string)

	TimeFormat() string
	SetTimeFormat(timeFormat string)
}
