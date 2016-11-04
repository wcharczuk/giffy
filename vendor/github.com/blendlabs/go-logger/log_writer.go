package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// DefaultBufferPoolSize is the default buffer pool size.
	DefaultBufferPoolSize = 1 << 8 // 256

	// DefaultTimeFormat is the default time format.
	DefaultTimeFormat = time.RFC3339
)

// NewLogWriterFromEnvironment initializes a log writer from the environment.
func NewLogWriterFromEnvironment() *LogWriter {
	return &LogWriter{
		Output:        NewStdOutMultiWriterFromEnvironment(),
		ErrorOutput:   NewStdErrMultiWriterFromEnvironment(),
		useAnsiColors: envFlagIsSet(EnvironmentVariableUseAnsiColors, true),
		showTimestamp: envFlagIsSet(EnvironmentVariableShowTimestamp, true),
		showLabel:     envFlagIsSet(EnvironmentVariableShowLabel, true),
		label:         os.Getenv(EnvironmentVariableLogLabel),
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// NewLogWriter returns a new writer.
func NewLogWriter(output io.Writer, optionalErrorOutput ...io.Writer) *LogWriter {
	agent := &LogWriter{
		Output:        NewSyncWriter(output),
		useAnsiColors: true,
		showTimestamp: true,
		showLabel:     false,
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
	if len(optionalErrorOutput) > 0 {
		agent.ErrorOutput = optionalErrorOutput[0]
	}
	return agent
}

// LogWriter handles outputting logging events to given writer streams.
type LogWriter struct {
	Output      io.Writer
	ErrorOutput io.Writer

	showTimestamp bool
	showLabel     bool
	useAnsiColors bool

	timeFormat string
	label      string

	bufferPool *BufferPool
}

// GetErrorOutput returns an io.Writer for the error stream.
func (wr *LogWriter) GetErrorOutput() io.Writer {
	if wr.ErrorOutput != nil {
		return wr.ErrorOutput
	}
	return wr.Output
}

// Colorize (optionally) applies a color to a string.
func (wr *LogWriter) Colorize(value string, color AnsiColorCode) string {
	if wr.useAnsiColors {
		return color.Apply(value)
	}
	return value
}

// ColorizeByStatusCode colorizes a string by a status code (green, yellow, red).
func (wr *LogWriter) ColorizeByStatusCode(statusCode int, value string) string {
	if wr.useAnsiColors {
		if statusCode >= http.StatusOK && statusCode < 300 { //the http 2xx range is ok
			return ColorGreen.Apply(value)
		} else if statusCode == http.StatusInternalServerError {
			return ColorRed.Apply(value)
		} else {
			return ColorYellow.Apply(value)
		}
	}
	return value
}

// GetTimestamp returns a new timestamp string.
func (wr *LogWriter) GetTimestamp(optionalTimeSource ...TimeSource) string {
	timeFormat := DefaultTimeFormat
	if len(wr.timeFormat) > 0 {
		timeFormat = wr.timeFormat
	}
	if len(optionalTimeSource) > 0 {
		return wr.Colorize(optionalTimeSource[0].UTCNow().Format(timeFormat), ColorGray)
	}
	return wr.Colorize(time.Now().UTC().Format(timeFormat), ColorGray)
}

// formatLabel returns the app name.
func (wr *LogWriter) formatLabel() string {
	return wr.Colorize(wr.label, ColorBlue)
}

// Printf writes to the output stream.
func (wr *LogWriter) Printf(format string, args ...interface{}) {
	wr.Fprintf(wr.Output, format, args...)
}

// PrintfWithTimeSource writes to the output stream, with a given timing source.
func (wr *LogWriter) PrintfWithTimeSource(ts TimeSource, format string, args ...interface{}) {
	wr.FprintfWithTimeSource(ts, wr.Output, format, args...)
}

// Errorf writes to the error output stream.
func (wr *LogWriter) Errorf(format string, args ...interface{}) {
	wr.Fprintf(wr.GetErrorOutput(), format, args...)
}

// ErrorfWithTimeSource writes to the error output stream, with a given timing source.
func (wr *LogWriter) ErrorfWithTimeSource(ts TimeSource, format string, args ...interface{}) {
	wr.FprintfWithTimeSource(ts, wr.GetErrorOutput(), format, args...)
}

// Write writes a binary blob to a given writer, and with a given timing source.
func (wr *LogWriter) Write(binary []byte) (int64, error) {
	return wr.WriteWithTimeSource(SystemClock, binary)
}

// WriteWithTimeSource writes a binary blob to a given writer, and with a given timing source.
func (wr *LogWriter) WriteWithTimeSource(ts TimeSource, binary []byte) (int64, error) {
	buf := wr.bufferPool.Get()
	defer wr.bufferPool.Put(buf)

	if wr.showTimestamp {
		buf.WriteString(wr.GetTimestamp(ts))
		buf.WriteRune(RuneSpace)
	}

	if wr.showLabel && len(wr.label) > 0 {
		buf.WriteString(wr.formatLabel())
		buf.WriteRune(RuneSpace)
	}

	buf.Write(binary)
	buf.WriteRune(RuneNewline)
	return buf.WriteTo(wr.Output)
}

// Fprintf writes a given string and args to a writer.
func (wr *LogWriter) Fprintf(w io.Writer, format string, args ...interface{}) {
	wr.FprintfWithTimeSource(SystemClock, w, format, args...)
}

// FprintfWithTimeSource writes a given string and args to a writer and with a given timing source.
func (wr *LogWriter) FprintfWithTimeSource(ts TimeSource, w io.Writer, format string, args ...interface{}) {
	if w == nil {
		return
	}
	if len(format) == 0 {
		return
	}
	message := fmt.Sprintf(format, args...)
	if len(message) == 0 {
		return
	}

	buf := wr.bufferPool.Get()
	defer wr.bufferPool.Put(buf)

	if wr.showTimestamp {
		buf.WriteString(wr.GetTimestamp(ts))
		buf.WriteRune(RuneSpace)
	}

	if wr.showLabel && len(wr.label) > 0 {
		buf.WriteString(wr.formatLabel())
		buf.WriteRune(RuneSpace)
	}

	buf.WriteString(message)
	buf.WriteRune(RuneNewline)
	buf.WriteTo(w)
}

// UseAnsiColors is a formatting option.
func (wr *LogWriter) UseAnsiColors() bool { return wr.useAnsiColors }

// SetUseAnsiColors sets a formatting option.
func (wr *LogWriter) SetUseAnsiColors(useAnsiColors bool) { wr.useAnsiColors = useAnsiColors }

// ShowTimestamp is a formatting option.
func (wr *LogWriter) ShowTimestamp() bool { return wr.showTimestamp }

// SetShowTimestamp sets a formatting option.
func (wr *LogWriter) SetShowTimestamp(showTimestamp bool) { wr.showTimestamp = showTimestamp }

// ShowLabel is a formatting option.
func (wr *LogWriter) ShowLabel() bool { return wr.showLabel }

// SetShowLabel sets a formatting option.
func (wr *LogWriter) SetShowLabel(showLabel bool) { wr.showLabel = showLabel }

// Label is a formatting option.
func (wr *LogWriter) Label() string { return wr.label }

// SetLabel sets a formatting option.
func (wr *LogWriter) SetLabel(label string) { wr.label = label }

// TimeFormat is a formatting option.
func (wr *LogWriter) TimeFormat() string { return wr.timeFormat }

// SetTimeFormat sets a formatting option.
func (wr *LogWriter) SetTimeFormat(timeFormat string) { wr.timeFormat = timeFormat }

// GetBuffer returns a leased buffer from the buffer pool.
func (wr *LogWriter) GetBuffer() *bytes.Buffer {
	return wr.bufferPool.Get()
}

// PutBuffer adds the leased buffer back to the pool.
// It Should be called in conjunction with `GetBuffer`.
func (wr *LogWriter) PutBuffer(buffer *bytes.Buffer) {
	wr.bufferPool.Put(buffer)
}
