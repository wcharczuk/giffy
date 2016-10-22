package logger

import (
	"net/http"
	"time"
)

// EventListener is a listener for a specific event as given by its flag.
type EventListener func(writer Logger, ts TimeSource, eventFlag EventFlag, state ...interface{})

// ErrorListener is a handler for error events.
type ErrorListener func(writer Logger, ts TimeSource, err error)

// NewErrorHandler returns a new handler for EventFatalError and EventError events.
func NewErrorHandler(errorHandler ErrorListener) EventListener {
	return func(writer Logger, ts TimeSource, eventFlag EventFlag, state ...interface{}) {
		if len(state) > 0 {
			if typedError, isTyped := state[0].(error); isTyped {
				errorHandler(writer, ts, typedError)
			}
		}
	}
}

// RequestListener is a listener for request events.
type RequestListener func(writer Logger, ts TimeSource, req *http.Request)

// NewRequestHandler returns a new handler for request events.
func NewRequestHandler(reqHandler RequestListener) EventListener {
	return func(writer Logger, ts TimeSource, eventFlag EventFlag, state ...interface{}) {
		if len(state) > 0 {
			if typedRequest, isTyped := state[0].(*http.Request); isTyped {
				reqHandler(writer, ts, typedRequest)
			}
		}
	}
}

// RequestCompleteListener is a listener for request events.
type RequestCompleteListener func(writer Logger, ts TimeSource, req *http.Request, statusCode, contentLengthBytes int, elapsed time.Duration)

// NewRequestCompleteHandler returns a new handler for request events.
func NewRequestCompleteHandler(reqCompleteHandler RequestCompleteListener) EventListener {
	return func(writer Logger, ts TimeSource, eventFlag EventFlag, state ...interface{}) {
		if len(state) < 3 {
			return
		}

		req, err := stateAsRequest(state[0])
		if err != nil {
			return
		}

		statusCode, err := stateAsInteger(state[1])
		if err != nil {
			return
		}

		contentLengthBytes, err := stateAsInteger(state[2])
		if err != nil {
			return
		}

		elapsed, err := stateAsDuration(state[3])
		if err != nil {
			return
		}

		reqCompleteHandler(writer, ts, req, statusCode, contentLengthBytes, elapsed)
	}
}

// RequestBodyListener is a listener for request bodies.
type RequestBodyListener func(writer Logger, ts TimeSource, body []byte)

// NewRequestBodyHandler returns a new handler for request body events.
func NewRequestBodyHandler(reqBodyHandler RequestBodyListener) EventListener {
	return func(writer Logger, ts TimeSource, eventFlag EventFlag, state ...interface{}) {
		if len(state) < 1 {
			return
		}
		body, err := stateAsBytes(state[0])
		if err != nil {
			return
		}
		reqBodyHandler(writer, ts, body)
	}
}
