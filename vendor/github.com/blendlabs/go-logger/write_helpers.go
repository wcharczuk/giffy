package logger

import (
	"bytes"
	"net/http"
	"strconv"
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
func TextWriteRequest(tf TextFormatter, buf *bytes.Buffer, wre *WebRequestEvent) {
	req := wre.Request()
	buf.WriteString(GetIP(req))
	buf.WriteRune(RuneSpace)
	buf.WriteString(tf.Colorize(req.Method, ColorBlue))
	buf.WriteRune(RuneSpace)
	buf.WriteString(req.URL.Path)
	buf.WriteRune(RuneSpace)
	buf.WriteString(tf.ColorizeByStatusCode(wre.StatusCode(), strconv.Itoa(wre.StatusCode())))
	buf.WriteRune(RuneSpace)
	buf.WriteString(wre.Elapsed().String())
	buf.WriteRune(RuneSpace)
	buf.WriteString(wre.ContentType())
	buf.WriteRune(RuneSpace)
	buf.WriteString(FormatFileSize(wre.ContentLength()))
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
func JSONWriteRequest(wre *WebRequestEvent) JSONObj {
	req := wre.Request()
	return JSONObj{
		"ip":              GetIP(req),
		"verb":            req.Method,
		"path":            req.URL.Path,
		"queryString":     req.URL.RawQuery,
		"host":            req.Host,
		"route":           wre.Route(),
		"contentLength":   wre.ContentLength(),
		"contentType":     wre.ContentType(),
		"contentEncoding": wre.ContentEncoding(),
		"statusCode":      wre.StatusCode(),
		JSONFieldElapsed:  Milliseconds(wre.Elapsed()),
	}
}
