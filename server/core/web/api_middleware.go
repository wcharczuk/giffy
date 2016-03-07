package web

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/blendlabs/connectivity/core/logging"
	"github.com/julienschmidt/httprouter"
)

type APIControllerAction func(*APIContext) *ServiceResponse

func APIActionHandler(action APIControllerAction) httprouter.Handle {
	return GZipped(Logged(MarshalAPIControllerAction(action)))
}

func MarshalAPIControllerAction(action APIControllerAction) httprouter.Handle {
	jsonMiddleware := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		responseResult := action(NewAPIContext(w, r, p))
		if responseResult.Meta.HttpCode == http.StatusNoContent {
			WriteNoContent(w)
		} else {
			_, writeErr := WriteJson(w, r, responseResult.Meta.HttpCode, responseResult)
			if writeErr != nil {
				logging.LogError(writeErr)
			}
		}
	}
	return jsonMiddleware
}

func Logged(handler httprouter.Handle) httprouter.Handle {
	logMiddleware := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		handler(outerWriter, r, p)
		logging.LogRequest(r, p, statusCode, end.Sub(start), contentLength, nil)
	}
	return logMiddleware
}

func GZipped(handler httprouter.Handle) httprouter.Handle {
	gzipMiddleware := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler(w, r, p)
			return
		}

		byte_storage := new(bytes.Buffer)
		gzipped_writer := NewGZipResponseWriter(byte_storage)
		defer gzipped_writer.Writer.Close()

		handler(gzipped_writer, r, p)
		flush_error := gzipped_writer.Writer.Flush()
		if flush_error != nil {
			logging.LogErrorMessageSimple(flush_error.Error())
		}
		gzipped_writer.Writer.Close()
		gzipped_writer.Headers.Set("Content-Encoding", "gzip")

		for header, header_values := range gzipped_writer.Headers {
			for _, header_value := range header_values {
				w.Header().Set(header, header_value)
			}
		}
		w.WriteHeader(gzipped_writer.StatusCode)
		w.Write(byte_storage.Bytes())
	}
	return gzipMiddleware
}
