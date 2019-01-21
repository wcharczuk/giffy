package request

import (
	"bytes"
	"fmt"
	"time"

	"github.com/blend/go-sdk/logger"
)

const (
	// Flag is a logger event flag.
	Flag logger.Flag = "request"
	// FlagResponse is a logger event flag.
	FlagResponse logger.Flag = "request.response"
)

// NewRequestListener creates a new request listener.
func NewRequestListener(listener func(Event)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(Event); isTyped {
			listener(typed)
		}
	}
}

// Event is a logger event for outgoing requests.
type Event struct {
	ts  time.Time
	req *Meta
}

// Flag returns the event flag.
func (re Event) Flag() logger.Flag {
	return Flag
}

// Timestamp returns the event timestamp.
func (re Event) Timestamp() time.Time {
	return re.ts
}

// Request returns the request meta.
func (re Event) Request() *Meta {
	return re.req
}

// WriteText writes an outgoing request as text to a given buffer.
func (re Event) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%s %s", re.req.Method, re.req.URL.String()))
}

// WriteJSON implements logger.JSONWritable.
func (re Event) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"req": re.req,
	}
}

// ResponseEvent is a response to outgoing requests.
type ResponseEvent struct {
	ts   time.Time
	req  *Meta
	res  *ResponseMeta
	body []byte
}

// Flag returns the event flag.
func (re ResponseEvent) Flag() logger.Flag {
	return FlagResponse
}

// Timestamp returns the event timestamp.
func (re ResponseEvent) Timestamp() time.Time {
	return re.ts
}

// Request returns the request meta.
func (re ResponseEvent) Request() *Meta {
	return re.req
}

// Response returns the response meta.
func (re ResponseEvent) Response() *ResponseMeta {
	return re.res
}

// Body returns the outgoing request body.
func (re ResponseEvent) Body() []byte {
	return re.body
}

// WriteText writes the event to a text writer.
func (re ResponseEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%s %s %s", re.req.Method, re.req.URL.String(), tf.ColorizeStatusCode(re.res.StatusCode)))
	if len(re.body) > 0 {
		buf.WriteRune(logger.RuneNewline)
		buf.Write(re.body)
	}
}

// WriteJSON implements logger.JSONWritable.
func (re ResponseEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"req":  re.req,
		"res":  re.res,
		"body": re.body,
	}
}
