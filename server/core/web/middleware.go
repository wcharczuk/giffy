package web

import (
	"net/http"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/core"
)

// ControllerResult is the result of a controller.
type ControllerResult interface {
	Render(*HTTPContext) error
}

// ControllerAction is the function signature for controller actions.
type ControllerAction func(*HTTPContext) ControllerResult

// ActionHandler takes an APIControllerAction and makes it an httprouter.Handle.
func ActionHandler(action ControllerAction) httprouter.Handle {
	return Render(action)
}

// NotFoundHandler is a handler for panics.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	Render(func(ctx *HTTPContext) ControllerResult {
		return ctx.NotFound()
	})(w, r, httprouter.Params{})
}

// PanicHandler is a handler for panics.
func PanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	Render(func(ctx *HTTPContext) ControllerResult {
		return ctx.InternalError(exception.Newf("panic: %v", err))
	})(w, r, httprouter.Params{})
}

// Render is the translation step from APIControllerAction to httprouter.Handle.
func Render(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			renderUncompressed(action, w, r, p)
		} else {
			renderCompressed(action, w, r, p)
		}
	}
}

func renderUncompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Vary", "Accept-Encoding")
	rw := NewResponseWriter(w)
	context := NewHTTPContext(rw, r, p)
	context.onRequestStart()

	context.Render(action(context))

	context.setStatusCode(rw.StatusCode)
	context.setContentLength(rw.ContentLength)

	context.onRequestEnd()
	context.LogRequest(core.RequestLogFormat)
}

func renderCompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	gzw := NewGZippedResponseWriter(w)
	defer gzw.Close()

	context := NewHTTPContext(gzw, r, p)
	context.onRequestStart()

	result := action(context)
	context.Render(result)

	gzw.Flush()

	context.setStatusCode(gzw.StatusCode)
	context.setContentLength(gzw.BytesWritten)
	context.onRequestEnd()
	context.LogRequest(core.RequestLogFormat)
}
