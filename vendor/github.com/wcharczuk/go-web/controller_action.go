package web

// ControllerMiddleware are steps that run in order before a given action.
type ControllerMiddleware func(ControllerAction) ControllerAction

// ControllerAction is the function signature for controller actions.
type ControllerAction func(*RequestContext) ControllerResult

// PanicControllerAction is a receiver for app.PanicHandler.
type PanicControllerAction func(*RequestContext, interface{}) ControllerResult
