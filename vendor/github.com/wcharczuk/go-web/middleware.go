package web

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
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
	context.LogRequest()
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
	context.LogRequest()
}
