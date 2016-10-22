package web

import "github.com/blendlabs/go-logger"

// DiagnosticsRequestCompleteHandler is a handler that takes a request context.
type DiagnosticsRequestCompleteHandler func(rc *RequestContext)

// NewDiagnosticsRequestCompleteHandler returns a binder for EventListener.
func NewDiagnosticsRequestCompleteHandler(handler DiagnosticsRequestCompleteHandler) logger.EventListener {
	return func(wr logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) < 1 {
			return
		}

		rc, isRequestContext := state[0].(*RequestContext)
		if !isRequestContext {
			return
		}

		handler(rc)
	}
}

// DiagnosticsErrorHandler is a handler that takes a request context.
type DiagnosticsErrorHandler func(rc *RequestContext, err error)

// NewDiagnosticsErrorHandler returns a binder for EventListener.
func NewDiagnosticsErrorHandler(handler DiagnosticsErrorHandler) logger.EventListener {
	return func(wr logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) < 2 {
			return
		}

		rc, isRequestContext := state[0].(*RequestContext)
		if !isRequestContext {
			return
		}

		err, isError := state[1].(error)
		if !isError {
			return
		}

		handler(rc, err)
	}
}
