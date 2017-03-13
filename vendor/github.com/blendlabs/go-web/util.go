package web

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net"
	"net/http"

	"github.com/blendlabs/go-exception"
	"github.com/julienschmidt/httprouter"
)

// NestMiddleware reads the middleware variadic args and organizes the calls recursively in the order they appear.
func NestMiddleware(action Action, middleware ...Middleware) Action {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b Middleware) Middleware {
		if b == nil {
			return a
		}
		return func(action Action) Action {
			return a(b(action))
		}
	}

	var metaAction Middleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// WriteNoContent writes http.StatusNoContent for a request.
func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte{})
	return nil
}

// WriteRawContent writes raw content for the request.
func WriteRawContent(w http.ResponseWriter, statusCode int, content []byte) error {
	w.WriteHeader(statusCode)
	_, err := w.Write(content)
	return exception.Wrap(err)
}

// WriteJSON marshalls an object to json.
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	err := enc.Encode(response)
	return exception.Wrap(err)
}

// WriteXML marshalls an object to json.
func WriteXML(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeXML)
	w.WriteHeader(statusCode)

	enc := xml.NewEncoder(w)
	err := enc.Encode(response)
	return exception.Wrap(err)
}

// DeserializeReaderAsJSON deserializes a post body as json to a given object.
func DeserializeReaderAsJSON(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	return exception.Wrap(json.NewDecoder(body).Decode(object))
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

func newHandleShim(app *App, handler Action) http.Handler {
	return &handleShim{action: handler, app: app}
}

type handleShim struct {
	action Action
	app    *App
}

func (hs handleShim) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.app.renderAction(hs.action)(w, r, httprouter.Params{})
}
