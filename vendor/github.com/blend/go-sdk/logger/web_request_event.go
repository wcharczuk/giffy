package logger

import (
	"bytes"
	"net/http"
	"time"
)

// these are compile time assertions
var (
	_ Event            = &WebRequestEvent{}
	_ EventHeadings    = &WebRequestEvent{}
	_ EventLabels      = &WebRequestEvent{}
	_ EventAnnotations = &WebRequestEvent{}
)

// NewWebRequestEvent creates a new web request event.
func NewWebRequestEvent(req *http.Request) *WebRequestEvent {
	return &WebRequestEvent{
		EventMeta: NewEventMeta(WebRequest),
		req:       req,
	}
}

// NewWebRequestStartEvent creates a new web request start event.
func NewWebRequestStartEvent(req *http.Request) *WebRequestEvent {
	return &WebRequestEvent{
		EventMeta: NewEventMeta(WebRequestStart),
		req:       req,
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
	*EventMeta

	req             *http.Request
	route           string
	statusCode      int
	contentLength   int64
	contentType     string
	contentEncoding string
	elapsed         time.Duration
	state           map[string]interface{}
}

// WithHeadings sets the headings.
func (e *WebRequestEvent) WithHeadings(headings ...string) *WebRequestEvent {
	e.headings = headings
	return e
}

// WithLabel sets a label on the event for later filtering.
func (e *WebRequestEvent) WithLabel(key, value string) *WebRequestEvent {
	e.AddLabelValue(key, value)
	return e
}

// WithAnnotation adds an annotation to the event.
func (e *WebRequestEvent) WithAnnotation(key, value string) *WebRequestEvent {
	e.AddAnnotationValue(key, value)
	return e
}

// WithFlag sets the event flag.
func (e *WebRequestEvent) WithFlag(flag Flag) *WebRequestEvent {
	e.flag = flag
	return e
}

// WithTimestamp sets the timestamp.
func (e *WebRequestEvent) WithTimestamp(ts time.Time) *WebRequestEvent {
	e.ts = ts
	return e
}

// WithRequest sets the request metadata.
func (e *WebRequestEvent) WithRequest(req *http.Request) *WebRequestEvent {
	e.req = req
	return e
}

// Request returns the request metadata.
func (e *WebRequestEvent) Request() *http.Request {
	return e.req
}

// WithStatusCode sets the status code.
func (e *WebRequestEvent) WithStatusCode(statusCode int) *WebRequestEvent {
	e.statusCode = statusCode
	return e
}

// StatusCode is the HTTP status code of the response.
func (e *WebRequestEvent) StatusCode() int {
	return e.statusCode
}

// WithContentLength sets the content length.
func (e *WebRequestEvent) WithContentLength(contentLength int64) *WebRequestEvent {
	e.contentLength = contentLength
	return e
}

// ContentLength is the size of the response.
func (e *WebRequestEvent) ContentLength() int64 {
	return e.contentLength
}

// WithContentType sets the content type.
func (e *WebRequestEvent) WithContentType(contentType string) *WebRequestEvent {
	e.contentType = contentType
	return e
}

// ContentType is the type of the response.
func (e *WebRequestEvent) ContentType() string {
	return e.contentType
}

// WithContentEncoding sets the content encoding.
func (e *WebRequestEvent) WithContentEncoding(contentEncoding string) *WebRequestEvent {
	e.contentEncoding = contentEncoding
	return e
}

// ContentEncoding is the encoding of the response.
func (e *WebRequestEvent) ContentEncoding() string {
	return e.contentEncoding
}

// WithRoute sets the mux route.
func (e *WebRequestEvent) WithRoute(route string) *WebRequestEvent {
	e.route = route
	return e
}

// Route is the mux route of the request.
func (e *WebRequestEvent) Route() string {
	return e.route
}

// WithElapsed sets the elapsed time.
func (e *WebRequestEvent) WithElapsed(elapsed time.Duration) *WebRequestEvent {
	e.elapsed = elapsed
	return e
}

// Elapsed is the duration of the request.
func (e *WebRequestEvent) Elapsed() time.Duration {
	return e.elapsed
}

// WithState sets the request state.
func (e *WebRequestEvent) WithState(state map[string]interface{}) *WebRequestEvent {
	e.state = state
	return e
}

// State returns the state of the request.
func (e *WebRequestEvent) State() map[string]interface{} {
	return e.state
}

// WriteText implements TextWritable.
func (e *WebRequestEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	if e.flag == WebRequestStart {
		TextWriteRequestStart(formatter, buf, e.req)
	} else {
		TextWriteRequest(formatter, buf, e.req, e.statusCode, e.contentLength, e.contentType, e.elapsed)
	}
}

// WriteJSON implements JSONWritable.
func (e *WebRequestEvent) WriteJSON() JSONObj {
	if e.flag == WebRequestStart {
		return JSONWriteRequestStart(e.req)
	}
	return JSONWriteRequest(e.req, e.statusCode, e.contentLength, e.contentType, e.contentEncoding, e.elapsed)
}
