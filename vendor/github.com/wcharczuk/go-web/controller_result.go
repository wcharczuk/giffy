package web

// ControllerResult is the result of a controller.
type ControllerResult interface {
	Render(*RequestContext) error
}
