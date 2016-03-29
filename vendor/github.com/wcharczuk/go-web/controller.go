package web

import "github.com/julienschmidt/httprouter"

// Controller is an interface for controller objects.
type Controller interface {
	Register(router *httprouter.Router)
}
