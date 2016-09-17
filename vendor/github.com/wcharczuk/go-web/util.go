package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/julienschmidt/httprouter"
)

const (
	//HeaderContentType is the content type header.
	HeaderContentType = "Content-Type"

	//HeaderConnection is the connection header.
	HeaderConnection = "Connection"

	//ConnectionKeepAlive is the keep-alive connection header value.
	ConnectionKeepAlive = "keep-alive"

	//ContentTypeJSON is the standard json content type.
	ContentTypeJSON = "application/json; charset=utf-8"
)

// NestMiddleware reads the middleware variadic args and organizes the calls recursively in the order they appear.
func NestMiddleware(action ControllerAction, middleware ...ControllerMiddleware) ControllerAction {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b ControllerMiddleware) ControllerMiddleware {
		if b == nil {
			return a
		}
		return func(action ControllerAction) ControllerAction {
			return a(b(action))
		}
	}

	var metaAction ControllerMiddleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// WriteNoContent writes http.StatusNoContent for a request.
func WriteNoContent(w http.ResponseWriter) (int, error) {
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
	return 0, nil
}

// WriteRawContent writes raw content for the request.
func WriteRawContent(w http.ResponseWriter, statusCode int, content []byte) (int, error) {
	w.WriteHeader(statusCode)
	count, err := w.Write(content)
	return count, exception.Wrap(err)
}

// WriteJSON marshalls an object to json.
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) (int, error) {
	bytes, err := json.Marshal(response)
	if err != nil {
		return 0, exception.Wrap(err)
	}

	if requestConnectionHeader := r.Header.Get(HeaderConnection); !util.IsEmpty(requestConnectionHeader) {
		if strings.ToLower(requestConnectionHeader) == ConnectionKeepAlive {
			w.Header().Set(HeaderConnection, ConnectionKeepAlive)
		}
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)

	count, err := w.Write(bytes)
	return count, exception.Wrap(err)
}

// DeserializeReaderAsJSON deserializes a post body as json to a given object.
func DeserializeReaderAsJSON(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return exception.Wrap(err)
	}
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	return exception.Wrap(decoder.Decode(object))
}

// LocalIP returns the local server ip.
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func newHandleShim(app *App, handler ControllerAction) http.Handler {
	return &handleShim{action: handler, app: app}
}

type handleShim struct {
	action ControllerAction
	app    *App
}

func (hs handleShim) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.app.renderAction(hs.action)(w, r, httprouter.Params{})
}
