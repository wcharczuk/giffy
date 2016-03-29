package web

var (
	// ProviderAPI Shared instances of the APIResultProvider
	ProviderAPI = NewAPIResultProvider()

	// ProviderView Shared instances of the ViewResultProvider
	ProviderView = NewViewResultProvider()
)

// HTTPResultProvider is the provider interface for results.
type HTTPResultProvider interface {
	InternalError(err error) ControllerResult
	BadRequest(message string) ControllerResult
	NotFound() ControllerResult
	NotAuthorized() ControllerResult
}
