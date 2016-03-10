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

// ControllerResult is the result of a controller.
type ControllerResult interface{}

// ControllerAction is the function signature for controller actions.
type ControllerAction func(*HTTPContext) ControllerResult

// ActionHandler takes an APIControllerAction and makes it an httprouter.Handle.
func ActionHandler(action ControllerAction) httprouter.Handle {
	return GZipped(RenderControllerAction(action))
}

// APINotFoundHandler is a handler for panics.
func APINotFoundHandler(w http.ResponseWriter, r *http.Request) {
	context := NewHTTPContext(w, r, httprouter.Params{})
	context.MarkRequestStart()

	response := context.NotFound()
	WriteJSON(w, r, context.StatusCode(), response)

	context.MarkRequestEnd()
	context.LogRequest(core.RequestLogFormat)
}

// APIPanicHandler is a handler for panics.
func APIPanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	context := NewHTTPContext(w, r, httprouter.Params{})
	context.MarkRequestStart()

	response := context.InternalError(exception.New(fmt.Sprintf("%v", err)))
	WriteJSON(w, r, context.StatusCode(), response)

	context.MarkRequestEnd()
	context.LogRequest(core.RequestLogFormat)
}

func isJSON(result ControllerResult) bool {
	_, isAPIResult := result.(*APIResult)
	return isAPIResult
}

func isRedirect(result ControllerResult) (*RedirectResult, bool) {
	typedResult, isTyped := result.(*RedirectResult)
	return typedResult, isTyped
}

// RenderAPIControllerAction is the translation step from APIControllerAction to httprouter.Handle.
func RenderAPIControllerAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		context := NewHTTPContext(w, r, p)
		context.MarkRequestStart()
		responseResult := action(context)
		bytesWritten := 0
		if isJSON(responseResult) {
			bytesWritten, _ = WriteJSON(w, r, context.StatusCode(), responseResult)
		} else if redirectResult, shouldRedirect := isRedirect(responseResult); shouldRedirect {

		} else if viewResult, isViewResult := isViewResult(responseResult); isViewResult {
			viewResult.Render(w)
		}

		context.SetContentLength(bytesWritten)
		context.MarkRequestEnd()
		context.LogRequest(core.RequestLogFormat)
	}
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
