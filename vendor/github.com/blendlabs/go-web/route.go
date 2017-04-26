package web

import (
	"database/sql"
	"fmt"
	"net/http"
)

// Handler is the most basic route handler.
type Handler func(http.ResponseWriter, *http.Request, *Route, RouteParameters, *sql.Tx)

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
