package logger

import (
	"bytes"
	"net/http"
	"strconv"
	"time"
)

// TextWriteRequestStart is a helper method to write request start events to a writer.
func TextWriteRequestStart(tf TextFormatter, buf *bytes.Buffer, req *http.Request) {
	buf.WriteString(GetIP(req))
	buf.WriteRune(RuneSpace)
	buf.WriteString(tf.Colorize(req.Method, ColorBlue))
	buf.WriteRune(RuneSpace)
	buf.WriteString(req.URL.Path)
}

// TextWriteRequest is a helper method to write request complete events to a writer.
func TextWriteRequest(tf TextFormatter, buf *bytes.Buffer, req *http.Request, statusCode int, contentLength int64, contentType string, elapsed time.Duration) {
	buf.WriteString(GetIP(req))
	buf.WriteRune(RuneSpace)
	buf.WriteString(tf.Colorize(req.Method, ColorBlue))
	buf.WriteRune(RuneSpace)
	buf.WriteString(req.URL.Path)
	buf.WriteRune(RuneSpace)
	buf.WriteString(tf.ColorizeByStatusCode(statusCode, strconv.Itoa(statusCode)))
	buf.WriteRune(RuneSpace)
	buf.WriteString(elapsed.String())
	buf.WriteRune(RuneSpace)
	buf.WriteString(contentType)
	buf.WriteRune(RuneSpace)
	buf.WriteString(FormatFileSize(contentLength))
}

// JSONWriteRequestStart marshals a request start as json.
func JSONWriteRequestStart(req *http.Request) JSONObj {
	return JSONObj{
		"ip":   GetIP(req),
		"verb": req.Method,
		"path": req.URL.Path,
		"host": req.Host,
	}
}

// JSONWriteRequest marshals a request as json.
func JSONWriteRequest(req *http.Request, statusCode int, contentLength int64, contentType, contentEncoding string, elapsed time.Duration) JSONObj {
	return JSONObj{
		"ip":              GetIP(req),
		"verb":            req.Method,
		"path":            req.URL.Path,
		"host":            req.Host,
		"contentLength":   contentLength,
		"contentType":     contentType,
		"contentEncoding": contentEncoding,
		"statusCode":      statusCode,
		JSONFieldElapsed:  Milliseconds(elapsed),
	}
}
