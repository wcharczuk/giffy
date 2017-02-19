package web

// ResultProvider is the provider interface for results.
type ResultProvider interface {
	InternalError(err error) Result
	BadRequest(message string) Result
	NotFound() Result
	NotAuthorized() Result
	Result(response interface{}) Result
}
