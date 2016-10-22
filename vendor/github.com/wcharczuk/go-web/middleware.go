package web

// APIProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func APIProviderAsDefault(action ControllerAction) ControllerAction {
	return func(context *RequestContext) ControllerResult {
		context.SetDefaultResultProvider(context.API())
		return action(context)
	}
}

// ViewProviderAsDefault sets the context.CurrrentProvider() equal to context.View().
func ViewProviderAsDefault(action ControllerAction) ControllerAction {
	return func(context *RequestContext) ControllerResult {
		context.SetDefaultResultProvider(context.API())
		return action(context)
	}
}
