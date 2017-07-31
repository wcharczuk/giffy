package web

import logger "github.com/blendlabs/go-logger"

const (
	// EventWebRequestStart is an aliased event flag.
	EventWebRequestStart = logger.EventWebRequestStart

	// EventWebRequest is an aliased event flag.
	EventWebRequest = logger.EventWebRequest

	// EventWebResponse is an aliased event flag.
	EventWebResponse = logger.EventWebResponse

	// EventWebRequestPostBody is an aliased event flag.
	EventWebRequestPostBody = logger.EventWebRequestPostBody

	// EventAppStart fires when the app is starting.
	EventAppStart = logger.EventFlag("web.app.start")

	// EventAppStartComplete fires after the app has started.
	EventAppStartComplete = logger.EventFlag("web.app.start.complete")

	// EventAppExit fires when an app exits.
	EventAppExit = logger.EventFlag("web.app.exit")
)

// RequestListener is a listener for `EventRequestStart` and `EventRequest` events.
type RequestListener func(*logger.Writer, logger.TimeSource, *Ctx)

// NewRequestListener creates a new logger.EventListener for `EventRequestStart` and `EventRequest` events.
func NewRequestListener(listener RequestListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) > 0 {
			if ctx, isCtx := state[0].(*Ctx); isCtx {
				listener(writer, ts, ctx)
			}
		}
	}
}

// ErrorListener is a listener for errors with an associated request context.
type ErrorListener func(*logger.Writer, logger.TimeSource, error, *Ctx)

// NewErrorListener returns a new error listener.
func NewErrorListener(listener ErrorListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) > 0 {
			if err, isError := state[0].(error); isError {

				if len(state) > 1 {
					if ctx, hasCtx := state[1].(*Ctx); hasCtx {
						listener(writer, ts, err, ctx)
						return
					}
				}
				listener(writer, ts, err, nil)
			}
		}
	}
}
