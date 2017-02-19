package web

import logger "github.com/blendlabs/go-logger"

// RequestListener is a listener for `EventRequestStart` and `EventRequest` events.
type RequestListener func(logger.Logger, logger.TimeSource, *Ctx)

// NewRequestListener creates a new logger.EventListener for `EventRequestStart` and `EventRequest` events.
func NewRequestListener(action RequestListener) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		action(writer, ts, state[0].(*Ctx))
	}
}
