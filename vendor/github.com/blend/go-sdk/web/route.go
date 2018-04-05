package web

import (
	"fmt"
	"net/http"
)

// Handler is the most basic route handler.
type Handler func(http.ResponseWriter, *http.Request, *Route, RouteParameters, State)

// WrapHandler wraps an http.Handler as a Handler.
func WrapHandler(handler http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request, _ *Route, _ RouteParameters, _ State) {
		handler.ServeHTTP(w, r)
	}
}

// PanicHandler is a handler for panics that also takes an error.
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

// Route is an entry in the route tree.
type Route struct {
	Handler
	Method string
	Path   string
	Params []string
}

// String returns a string representation of the route.
// Namely: Method_Path
func (r Route) String() string {
	return fmt.Sprintf("%s_%s", r.Method, r.Path)
}
