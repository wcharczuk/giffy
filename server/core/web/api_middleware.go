package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/core"
)

// APIControllerAction is the function signature for controller actions.
type APIControllerAction func(*APIContext) *ServiceResponse

// APIActionHandler takes an APIControllerAction and makes it an httprouter.Handle.
func APIActionHandler(action APIControllerAction) httprouter.Handle {
	return GZipped(MarshalAPIControllerAction(Logged(action)))
}

// APINotFoundHandler is a handler for panics.
func APINotFoundHandler(w http.ResponseWriter, r *http.Request) {
	context := NewAPIContext(w, r, httprouter.Params{})
	response := context.NotFound()
	WriteJSON(w, r, context.StatusCode(), response)
}

// APIPanicHandler is a handler for panics.
func APIPanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	context := NewAPIContext(w, r, httprouter.Params{})
	response := context.InternalError(exception.New(fmt.Sprintf("%v", err)))
	WriteJSON(w, r, context.StatusCode(), response)
}

// MarshalAPIControllerAction is the translation step from APIControllerAction to httprouter.Handle.
func MarshalAPIControllerAction(action APIControllerAction) httprouter.Handle {
	jsonMiddleware := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		context := NewAPIContext(w, r, p)
		responseResult := action(context)
		WriteJSON(w, r, context.StatusCode(), responseResult)
	}
	return jsonMiddleware
}

// Logged is a middleware step that logs requests.
func Logged(handler APIControllerAction) APIControllerAction {
	logMiddleware := func(ctx *APIContext) *ServiceResponse {
		ctx.MarkRequestStart()
		response := handler(ctx)
		ctx.MarkRequestEnd()
		ctx.LogRequest(core.RequestLogFormat)
		return response
	}
	return logMiddleware
}

// GZipped is a middleware step that compresses the response output.
func GZipped(handler httprouter.Handle) httprouter.Handle {
	gzipMiddleware := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler(w, r, p)
			return
		}

		byteStorage := new(bytes.Buffer)
		gzippedWriter := NewGZippedResponseWriter(byteStorage)
		defer gzippedWriter.Writer.Close()

		handler(gzippedWriter, r, p)
		err := gzippedWriter.Writer.Flush()
		if err != nil {
			LogError(err)
		}
		gzippedWriter.Writer.Close()
		gzippedWriter.Headers.Set("Content-Encoding", "gzip")

		for header, headerValues := range gzippedWriter.Headers {
			for _, headerValue := range headerValues {
				w.Header().Set(header, headerValue)
			}
		}
		w.WriteHeader(gzippedWriter.StatusCode)
		w.Write(byteStorage.Bytes())
	}
	return gzipMiddleware
}
