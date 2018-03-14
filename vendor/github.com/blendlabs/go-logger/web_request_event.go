package logger

import (
	"bytes"
	"net/http"
	"time"
)

// NewWebRequestEvent creates a new web request event.
func NewWebRequestEvent(req *http.Request) *WebRequestEvent {
	return &WebRequestEvent{
		flag: WebRequest,
		ts:   time.Now().UTC(),
		req:  req,
	}
}

// NewWebRequestStartEvent creates a new web request start event.
func NewWebRequestStartEvent(req *http.Request) *WebRequestEvent {
	return &WebRequestEvent{
		flag: WebRequestStart,
		ts:   time.Now().UTC(),
		req:  req,
	}
}

// NewWebRequestEventListener returns a new web request event listener.
func NewWebRequestEventListener(listener func(*WebRequestEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(*WebRequestEvent); isTyped {
			listener(typed)
		}
	}
}

// WebRequestEvent is an event type for http responses.
type WebRequestEvent struct {
	flag Flag
	ts   time.Time
	req  *http.Request

	label           string
	route           string
	statusCode      int
	contentLength   int64
	contentType     string
	contentEncoding string
	elapsed         time.Duration
	state           map[string]interface{}
}

// WithFlag sets the event flag.
func (wre *WebRequestEvent) WithFlag(flag Flag) *WebRequestEvent {
	wre.flag = flag
	return wre
}

// Flag returns the event flag.
func (wre WebRequestEvent) Flag() Flag {
	return wre.flag
}

// WithTimestamp sets the timestamp.
func (wre *WebRequestEvent) WithTimestamp(ts time.Time) *WebRequestEvent {
	wre.ts = ts
	return wre
}

// Timestamp returns the event timestamp.
func (wre WebRequestEvent) Timestamp() time.Time {
	return wre.ts
}

// WithLabel sets the label.
func (wre *WebRequestEvent) WithLabel(label string) *WebRequestEvent {
	wre.label = label
	return wre
}

// Label returns the label.
func (wre WebRequestEvent) Label() string {
	return wre.label
}

// WithRequest sets the request metadata.
func (wre *WebRequestEvent) WithRequest(req *http.Request) *WebRequestEvent {
	wre.req = req
	return wre
}

// Request returns the request metadata.
func (wre WebRequestEvent) Request() *http.Request {
	return wre.req
}

// WithStatusCode sets the status code.
func (wre *WebRequestEvent) WithStatusCode(statusCode int) *WebRequestEvent {
	wre.statusCode = statusCode
	return wre
}

// StatusCode is the HTTP status code of the response.
func (wre WebRequestEvent) StatusCode() int {
	return wre.statusCode
}

// WithContentLength sets the content length.
func (wre *WebRequestEvent) WithContentLength(contentLength int64) *WebRequestEvent {
	wre.contentLength = contentLength
	return wre
}

// ContentLength is the size of the response.
func (wre WebRequestEvent) ContentLength() int64 {
	return wre.contentLength
}

// WithContentType sets the content type.
func (wre *WebRequestEvent) WithContentType(contentType string) *WebRequestEvent {
	wre.contentType = contentType
	return wre
}

// ContentType is the type of the response.
func (wre WebRequestEvent) ContentType() string {
	return wre.contentType
}

// WithContentEncoding sets the content encoding.
func (wre *WebRequestEvent) WithContentEncoding(contentEncoding string) *WebRequestEvent {
	wre.contentEncoding = contentEncoding
	return wre
}

// ContentEncoding is the encoding of the response.
func (wre WebRequestEvent) ContentEncoding() string {
	return wre.contentEncoding
}

// WithRoute sets the mux route.
func (wre *WebRequestEvent) WithRoute(route string) *WebRequestEvent {
	wre.route = route
	return wre
}

// Route is the mux route of the request.
func (wre WebRequestEvent) Route() string {
	return wre.route
}

// WithElapsed sets the elapsed time.
func (wre *WebRequestEvent) WithElapsed(elapsed time.Duration) *WebRequestEvent {
	wre.elapsed = elapsed
	return wre
}

// Elapsed is the duration of the request.
func (wre WebRequestEvent) Elapsed() time.Duration {
	return wre.elapsed
}

// WithState sets the request state.
func (wre *WebRequestEvent) WithState(state map[string]interface{}) *WebRequestEvent {
	wre.state = state
	return wre
}

// State returns the state of the request.
func (wre WebRequestEvent) State() time.Duration {
	return wre.elapsed
}

// WriteText implements TextWritable.
func (wre *WebRequestEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	if wre.flag == WebRequestStart {
		TextWriteRequestStart(formatter, buf, wre.Request())
	} else {
		TextWriteRequest(formatter, buf, wre)
	}
}

// WriteJSON implements JSONWritable.
func (wre *WebRequestEvent) WriteJSON() JSONObj {
	if wre.flag == WebRequestStart {
		return JSONWriteRequestStart(wre.req)
	}
	return JSONWriteRequest(wre)
}
