package web

// ResultProvider is the provider interface for results.
type ResultProvider interface {
	InternalError(err error) Result
	BadRequest(err error) Result
	NotFound() Result
	NotAuthorized() Result
}
