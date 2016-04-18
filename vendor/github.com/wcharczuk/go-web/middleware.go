package web

// InjectAPIProvider sets the context.CurrrentProvider() equal to context.API().
func InjectAPIProvider(action ControllerAction) ControllerAction {
	return func(context *RequestContext) ControllerResult {
		context.SetCurrentProvider(context.API())
		return action(context)
	}
}

// InjectViewProvider sets the context.CurrrentProvider() equal to context.View().
func InjectViewProvider(action ControllerAction) ControllerAction {
	return func(context *RequestContext) ControllerResult {
		context.SetCurrentProvider(context.API())
		return action(context)
	}
}
