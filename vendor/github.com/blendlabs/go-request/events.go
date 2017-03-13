package request

import (
	"fmt"
	"strconv"

	logger "github.com/blendlabs/go-logger"
)

const (
	// Event is a diagnostics agent event flag.
	Event logger.EventFlag = "request"
	// EventResponse is a diagnostics agent event flag.
	EventResponse logger.EventFlag = "request.response"
)

// NewOutgoingListener creates a new logger handler for `EventFlagOutgoingResponse` events.
func NewOutgoingListener(handler func(writer logger.Logger, ts logger.TimeSource, req *HTTPRequestMeta)) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		handler(writer, ts, state[0].(*HTTPRequestMeta))
	}
}

// WriteOutgoingRequest is a helper method to write outgoing request events to a logger writer.
func WriteOutgoingRequest(writer logger.Logger, ts logger.TimeSource, req *HTTPRequestMeta) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)
	buffer.WriteString(writer.Colorize(string(Event), logger.ColorGreen))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(fmt.Sprintf("%s %s", req.Verb, req.URL.String()))
	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// WriteOutgoingRequestBody is a helper method to write outgoing request bodies to a logger writer.
func WriteOutgoingRequestBody(writer logger.Logger, ts logger.TimeSource, req *HTTPRequestMeta) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)
	buffer.WriteString(writer.Colorize(string(Event), logger.ColorGreen))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString("request body")
	buffer.WriteRune(logger.RuneNewline)
	buffer.Write(req.Body)
	writer.WriteWithTimeSource(ts, buffer.Bytes())
}

// NewOutgoingResponseListener creates a new logger handler for `EventFlagOutgoingResponse` events.
func NewOutgoingResponseListener(handler func(writer logger.Logger, ts logger.TimeSource, req *HTTPRequestMeta, res *HTTPResponseMeta, body []byte)) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		handler(writer, ts, state[0].(*HTTPRequestMeta), state[1].(*HTTPResponseMeta), state[2].([]byte))
	}
}

// WriteOutgoingRequestResponse is a helper method to write outgoing request response events to a logger writer.
func WriteOutgoingRequestResponse(writer logger.Logger, ts logger.TimeSource, req *HTTPRequestMeta, res *HTTPResponseMeta, body []byte) {
	buffer := writer.GetBuffer()
	defer writer.PutBuffer(buffer)
	buffer.WriteString(writer.Colorize(string(EventResponse), logger.ColorGreen))
	buffer.WriteRune(logger.RuneSpace)
	buffer.WriteString(fmt.Sprintf("%s %s %s", writer.ColorizeByStatusCode(res.StatusCode, strconv.Itoa(res.StatusCode)), req.Verb, req.URL.String()))
	buffer.WriteRune(logger.RuneNewline)
	buffer.Write(body)
	writer.WriteWithTimeSource(ts, buffer.Bytes())
}
