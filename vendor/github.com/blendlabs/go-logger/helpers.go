package logger

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// WriteEventf is a helper for creating new logging messasges.
func WriteEventf(writer Logger, ts TimeSource, label string, color AnsiColorCode, format string, args ...interface{}) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)

	buffer.WriteString(writer.Colorize(label, color))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(fmt.Sprintf(format, args...))
	buffer.WriteRune(RuneSpace)

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// WriteRequest is a helper method to write request start events to a writer.
func WriteRequest(writer Logger, ts TimeSource, req *http.Request) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)

	buffer.WriteString(writer.Colorize("Request", ColorGreen))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(GetIP(req))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(writer.Colorize(req.Method, ColorBlue))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(req.URL.Path)

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// WriteRequestComplete is a helper method to write request complete events to a writer.
func WriteRequestComplete(writer Logger, ts TimeSource, req *http.Request, statusCode, contentLengthBytes int, elapsed time.Duration) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)

	buffer.WriteString(writer.Colorize("Request Complete", ColorGreen))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(GetIP(req))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(writer.Colorize(req.Method, ColorBlue))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(req.URL.Path)
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(writer.ColorizeByStatusCode(statusCode, strconv.Itoa(statusCode)))
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(elapsed.String())
	buffer.WriteRune(RuneSpace)
	buffer.WriteString(FormatFileSize(contentLengthBytes))

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// WriteRequestBody is a helper method to write request start events to a writer.
func WriteRequestBody(writer Logger, ts TimeSource, body []byte) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)
	buffer.WriteString(writer.Colorize("Request Body", ColorGreen))
	buffer.WriteRune(RuneSpace)
	buffer.Write(body)

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// WriteResponseBody is a helper method to write request start events to a writer.
func WriteResponseBody(writer Logger, ts TimeSource, body []byte) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)
	buffer.WriteString(writer.Colorize("Response", ColorGreen))
	buffer.WriteRune(RuneSpace)
	buffer.Write(body)

	writer.WriteWithTimeSource(ts, buffer.Bytes())
}
