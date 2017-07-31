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

	// DefaultWriterUseAnsiColors is a default setting for writers.
	DefaultWriterUseAnsiColors = true
	// DefaultWriterShowTimestamp is a default setting for writers.
	DefaultWriterShowTimestamp = true
	// DefaultWriterShowLabel is a default setting for writers.
	DefaultWriterShowLabel = false
)

// NewWriter returns a new writer with combined standard and error outputs.
func NewWriter(output io.Writer) *Writer {
	agent := &Writer{
		Output:        NewSyncOutput(output),
		useAnsiColors: DefaultWriterUseAnsiColors,
		showTimestamp: DefaultWriterShowTimestamp,
		showLabel:     DefaultWriterShowLabel,
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
	return agent
}

// NewWriterWithError returns a new writer with a dedicated error output.
func NewWriterWithError(output, errorOutput io.Writer) *Writer {
	agent := &Writer{
		Output:        NewSyncOutput(output),
		ErrorOutput:   NewSyncOutput(errorOutput),
		useAnsiColors: DefaultWriterUseAnsiColors,
		showTimestamp: DefaultWriterShowTimestamp,
		showLabel:     DefaultWriterShowLabel,
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
	return agent
}

// NewWriterFromEnvironment initializes a log writer from the environment.
func NewWriterFromEnvironment() *Writer {
	return &Writer{
		Output:        NewMultiOutputFromEnvironment(),
		ErrorOutput:   NewErrorMultiOutputFromEnvironment(),
		useAnsiColors: envFlagIsSet(EnvironmentVariableUseAnsiColors, DefaultWriterUseAnsiColors),
		showTimestamp: envFlagIsSet(EnvironmentVariableShowTimestamp, DefaultWriterShowTimestamp),
		showLabel:     envFlagIsSet(EnvironmentVariableShowLabel, DefaultWriterShowLabel),
		label:         os.Getenv(EnvironmentVariableLogLabel),
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// NewWriterToFile creates a new writer that writes to stdout + stderr and a file.
func NewWriterToFile(path string) *Writer {
	fileoutput, err := NewFileOutputWithDefaults(path)
	if err != nil {
		panic(err)
	}
	return &Writer{
		Output:        NewMultiOutput(NewSyncOutput(os.Stdout), fileoutput),
		useAnsiColors: envFlagIsSet(EnvironmentVariableUseAnsiColors, DefaultWriterUseAnsiColors),
		showTimestamp: envFlagIsSet(EnvironmentVariableShowTimestamp, DefaultWriterShowTimestamp),
		showLabel:     envFlagIsSet(EnvironmentVariableShowLabel, DefaultWriterShowLabel),
		label:         os.Getenv(EnvironmentVariableLogLabel),
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// NewWriterToFileWithError creates a new writer that writes to stdout + stderr and a file.
func NewWriterToFileWithError(output, errorOutput string) *Writer {
	fileOutput, err := NewFileOutputWithDefaults(output)
	if err != nil {
		panic(err)
	}

	fileErrorOutput, err := NewFileOutputWithDefaults(errorOutput)
	if err != nil {
		panic(err)
	}

	return &Writer{
		Output:        NewMultiOutput(NewSyncOutput(os.Stdout), fileOutput),
		ErrorOutput:   NewMultiOutput(NewSyncOutput(os.Stderr), fileErrorOutput),
		useAnsiColors: envFlagIsSet(EnvironmentVariableUseAnsiColors, DefaultWriterUseAnsiColors),
		showTimestamp: envFlagIsSet(EnvironmentVariableShowTimestamp, DefaultWriterShowTimestamp),
		showLabel:     envFlagIsSet(EnvironmentVariableShowLabel, DefaultWriterShowLabel),
		label:         os.Getenv(EnvironmentVariableLogLabel),
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// Writer handles outputting logging events to given writer streams.
type Writer struct {
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
func (wr *Writer) GetErrorOutput() io.Writer {
	if wr.ErrorOutput != nil {
		return wr.ErrorOutput
	}
	return wr.Output
}

// Colorize (optionally) applies a color to a string.
func (wr *Writer) Colorize(value string, color AnsiColorCode) string {
	if wr.useAnsiColors {
		return color.Apply(value)
	}
	return value
}

// ColorizeByStatusCode colorizes a string by a status code (green, yellow, red).
func (wr *Writer) ColorizeByStatusCode(statusCode int, value string) string {
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
func (wr *Writer) GetTimestamp(optionalTimeSource ...TimeSource) string {
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
func (wr *Writer) formatLabel() string {
	return wr.Colorize(wr.label, ColorBlue)
}

// Printf writes to the output stream.
func (wr *Writer) Printf(format string, args ...interface{}) (int64, error) {
	return wr.Fprintf(wr.Output, format, args...)
}

// PrintfWithTimeSource writes to the output stream, with a given timing source.
func (wr *Writer) PrintfWithTimeSource(ts TimeSource, format string, args ...interface{}) (int64, error) {
	return wr.FprintfWithTimeSource(ts, wr.Output, format, args...)
}

// Errorf writes to the error output stream.
func (wr *Writer) Errorf(format string, args ...interface{}) (int64, error) {
	return wr.Fprintf(wr.GetErrorOutput(), format, args...)
}

// ErrorfWithTimeSource writes to the error output stream, with a given timing source.
func (wr *Writer) ErrorfWithTimeSource(ts TimeSource, format string, args ...interface{}) (int64, error) {
	return wr.FprintfWithTimeSource(ts, wr.GetErrorOutput(), format, args...)
}

// Write writes a binary blob to a given writer, and with a given timing source.
func (wr *Writer) Write(binary []byte) (int64, error) {
	return wr.WriteWithTimeSource(SystemClock, binary)
}

// WriteWithTimeSource writes a binary blob to a given writer, and with a given timing source.
func (wr *Writer) WriteWithTimeSource(ts TimeSource, binary []byte) (int64, error) {
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
func (wr *Writer) Fprintf(w io.Writer, format string, args ...interface{}) (int64, error) {
	return wr.FprintfWithTimeSource(SystemClock, w, format, args...)
}

// FprintfWithTimeSource writes a given string and args to a writer and with a given timing source.
func (wr *Writer) FprintfWithTimeSource(ts TimeSource, w io.Writer, format string, args ...interface{}) (int64, error) {
	if w == nil {
		return 0, nil
	}
	if len(format) == 0 {
		return 0, nil
	}
	message := fmt.Sprintf(format, args...)
	if len(message) == 0 {
		return 0, nil
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
	return buf.WriteTo(w)
}

// UseAnsiColors is a formatting option.
func (wr *Writer) UseAnsiColors() bool { return wr.useAnsiColors }

// SetUseAnsiColors sets a formatting option.
func (wr *Writer) SetUseAnsiColors(useAnsiColors bool) { wr.useAnsiColors = useAnsiColors }

// ShowTimestamp is a formatting option.
func (wr *Writer) ShowTimestamp() bool { return wr.showTimestamp }

// SetShowTimestamp sets a formatting option.
func (wr *Writer) SetShowTimestamp(showTimestamp bool) { wr.showTimestamp = showTimestamp }

// ShowLabel is a formatting option.
func (wr *Writer) ShowLabel() bool { return wr.showLabel }

// SetShowLabel sets a formatting option.
func (wr *Writer) SetShowLabel(showLabel bool) { wr.showLabel = showLabel }

// Label is a formatting option.
func (wr *Writer) Label() string { return wr.label }

// SetLabel sets a formatting option.
func (wr *Writer) SetLabel(label string) { wr.label = label }

// TimeFormat is a formatting option.
func (wr *Writer) TimeFormat() string { return wr.timeFormat }

// SetTimeFormat sets a formatting option.
func (wr *Writer) SetTimeFormat(timeFormat string) { wr.timeFormat = timeFormat }

// GetBuffer returns a leased buffer from the buffer pool.
func (wr *Writer) GetBuffer() *bytes.Buffer {
	return wr.bufferPool.Get()
}

// PutBuffer adds the leased buffer back to the pool.
// It Should be called in conjunction with `GetBuffer`.
func (wr *Writer) PutBuffer(buffer *bytes.Buffer) {
	wr.bufferPool.Put(buffer)
}

// Close closes the writer, free-ing underlying resources.
func (wr *Writer) Close() (err error) {
	if wr.Output != nil {
		if closer, isCloser := wr.Output.(io.Closer); isCloser {
			err = closer.Close()
			if err != nil {
				return
			}
		}
		wr.Output = nil
	}
	if wr.ErrorOutput != nil {
		if closer, isCloser := wr.ErrorOutput.(io.Closer); isCloser {
			err = closer.Close()
			if err != nil {
				return
			}
		}
		wr.ErrorOutput = nil
	}

	wr.bufferPool = nil
	return
}
