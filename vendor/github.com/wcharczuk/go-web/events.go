package web

import logger "github.com/blendlabs/go-logger"

const (
	// EventRequest is a logger event.
	EventRequest = logger.EventWebRequest

	// EventRequestStart is a logger event.
	EventRequestStart = logger.EventWebRequestStart

	// EventRequestPostBody is a logger event.
	EventRequestPostBody = logger.EventWebRequestPostBody

	// EventResponse is a logger event.
	EventResponse = logger.EventWebResponse
)

// RequestListener is a listener for `EventRequestStart` and `EventRequest` events.
type RequestListener func(logger.Logger, logger.TimeSource, *RequestContext)

// NewRequestListener creates a new logger.EventListener for `EventRequestStart` and `EventRequest` events.
func NewRequestListener(action RequestListener) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		action(writer, ts, state[0].(*RequestContext))
	}
}
