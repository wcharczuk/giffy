package web

// APIProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func APIProviderAsDefault(action Action) Action {
	return func(context *Ctx) Result {
		context.SetDefaultResultProvider(context.API())
		return action(context)
	}
}

// ViewProviderAsDefault sets the context.CurrrentProvider() equal to context.View().
func ViewProviderAsDefault(action Action) Action {
	return func(context *Ctx) Result {
		context.SetDefaultResultProvider(context.View())
		return action(context)
	}
}

// JSONProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func JSONProviderAsDefault(action Action) Action {
	return func(context *Ctx) Result {
		context.SetDefaultResultProvider(context.JSON())
		return action(context)
	}
}

// XMLProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func XMLProviderAsDefault(action Action) Action {
	return func(context *Ctx) Result {
		context.SetDefaultResultProvider(context.XML())
		return action(context)
	}
}
