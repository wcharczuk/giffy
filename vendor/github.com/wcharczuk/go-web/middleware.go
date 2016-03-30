package web

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// ControllerResult is the result of a controller.
type ControllerResult interface {
	Render(*RequestContext) error
}

// ControllerAction is the function signature for controller actions.
type ControllerAction func(*RequestContext) ControllerResult

// PanicControllerAction is a receiver for app.PanicHandler.
type PanicControllerAction func(*RequestContext, interface{}) ControllerResult

// ActionHandler takes an APIControllerAction and makes it an httprouter.Handle.
func ActionHandler(action ControllerAction) httprouter.Handle {
	return RenderAction(action)
}

// RenderAction is the translation step from APIControllerAction to httprouter.Handle.
func RenderAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			renderUncompressed(action, w, r, parseParams(p))
		} else {
			renderCompressed(action, w, r, parseParams(p))
		}
	}
}

func parseParams(p httprouter.Params) RouteParameters {
	rp := NewRouteParameters()
	for _, pv := range p {
		rp[pv.Key] = pv.Value
	}
	return rp
}

func renderUncompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Vary", "Accept-Encoding")
	rw := NewResponseWriter(w)
	context := NewRequestContext(rw, r, p)
	context.onRequestStart()
	context.Render(action(context))
	context.setStatusCode(rw.StatusCode)
	context.setContentLength(rw.ContentLength)
	context.onRequestEnd()
	context.LogRequest()
}

func renderCompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	gzw := NewGZippedResponseWriter(w)
	defer gzw.Close()
	context := NewRequestContext(gzw, r, p)
	context.onRequestStart()
	result := action(context)
	context.Render(result)
	gzw.Flush()
	context.setStatusCode(gzw.StatusCode)
	context.setContentLength(gzw.BytesWritten)
	context.onRequestEnd()
	context.LogRequest()
}
