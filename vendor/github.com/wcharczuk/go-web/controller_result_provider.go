package web

const (
	// Unknown is an unknown provider.
	ProviderUnknown = 0
	// API is the api provider.
	ProviderAPI = 1
	//View is the view provider.
	ProviderView = 2
)

// ControllerResultProvider is the provider interface for results.
type ControllerResultProvider interface {
	InternalError(err error) ControllerResult
	BadRequest(message string) ControllerResult
	NotFound() ControllerResult
	NotAuthorized() ControllerResult
}
