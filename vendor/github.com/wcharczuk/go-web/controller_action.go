package web

// ControllerAction is the function signature for controller actions.
type ControllerAction func(*RequestContext) ControllerResult

// PanicControllerAction is a receiver for app.PanicHandler.
type PanicControllerAction func(*RequestContext, interface{}) ControllerResult
