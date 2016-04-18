package web

import "net/http"

// ControllerResult is the result of a controller.
type ControllerResult interface {
	Render(w http.ResponseWriter, r *http.Request) error
}
