package web

import "github.com/blendlabs/go-logger"

// DiagnosticsRequestCompleteHandler is a handler that takes a request context.
type DiagnosticsRequestCompleteHandler func(ctx *Ctx)

// NewDiagnosticsRequestCompleteHandler returns a binder for EventListener.
func NewDiagnosticsRequestCompleteHandler(handler DiagnosticsRequestCompleteHandler) logger.EventListener {
	return func(wr *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) < 1 {
			return
		}

		ctx, isCtx := state[0].(*Ctx)
		if !isCtx {
			return
		}

		handler(ctx)
	}
}

// DiagnosticsErrorHandler is a handler that takes a request context.
type DiagnosticsErrorHandler func(ctx *Ctx, err error)

// NewDiagnosticsErrorHandler returns a binder for EventListener.
func NewDiagnosticsErrorHandler(handler DiagnosticsErrorHandler) logger.EventListener {
	return func(wr *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) < 2 {
			return
		}

		ctx, isCtx := state[0].(*Ctx)
		if !isCtx {
			return
		}

		err, isError := state[1].(error)
		if !isError {
			return
		}

		handler(ctx, err)
	}
}
