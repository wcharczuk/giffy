package web

// ControllerResultProvider is the provider interface for results.
type ControllerResultProvider interface {
	InternalError(err error) ControllerResult
	BadRequest(message string) ControllerResult
	NotFound() ControllerResult
	NotAuthorized() ControllerResult
}
