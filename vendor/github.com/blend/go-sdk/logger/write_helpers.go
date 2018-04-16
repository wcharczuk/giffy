package logger

import (
	"bytes"
	"net/http"
	"strconv"
	"time"
)

// FormatFileSize returns a string representation of a file size in bytes.
func FormatFileSize(sizeBytes int64) string {
	if sizeBytes >= 1<<30 {
		return strconv.FormatInt(sizeBytes/Gigabyte, 10) + "gb"
	} else if sizeBytes >= 1<<20 {
		return strconv.FormatInt(sizeBytes/Megabyte, 10) + "mb"
	} else if sizeBytes >= 1<<10 {
		return strconv.FormatInt(sizeBytes/Kilobyte, 10) + "kb"
	}
	return strconv.FormatInt(sizeBytes, 10)
}

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
