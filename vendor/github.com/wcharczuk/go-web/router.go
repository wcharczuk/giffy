package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Router is the interface any router implementations need to adhere to.
type Router interface {
	GET(path string, handler ControllerAction)
	OPTIONS(path string, handler ControllerAction)
	HEAD(path string, handler ControllerAction)
	POST(path string, handler ControllerAction)
	PUT(path string, handler ControllerAction)
	DELETE(path string, handler ControllerAction)

	Static(path string, root http.FileSystem)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// NewRouter returns a new router.
func NewRouter() Router {
	return &httpRouter{router: httprouter.New()}
}

type httpRouter struct {
	router *httprouter.Router
}

func (r *httpRouter) GET(path string, handler ControllerAction) {
	r.router.GET(path, ActionHandler(handler))
}

func (r *httpRouter) OPTIONS(path string, handler ControllerAction) {
	r.router.OPTIONS(path, ActionHandler(handler))
}

func (r *httpRouter) HEAD(path string, handler ControllerAction) {
	r.router.HEAD(path, ActionHandler(handler))
}

func (r *httpRouter) PUT(path string, handler ControllerAction) {
	r.router.PUT(path, ActionHandler(handler))
}

func (r *httpRouter) POST(path string, handler ControllerAction) {
	r.router.POST(path, ActionHandler(handler))
}

func (r *httpRouter) DELETE(path string, handler ControllerAction) {
	r.router.DELETE(path, ActionHandler(handler))
}

func (r *httpRouter) Static(path string, root http.FileSystem) {
	r.router.ServeFiles(path, root)
}

func (r *httpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
