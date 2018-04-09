package web

import (
	"github.com/blend/go-sdk/logger"
)

var noContent = &NoContentResult{}

// NoContentResultProvider is a provider that returns `http.StatusNoContent`
// for all responses.
type NoContentResultProvider struct {
	log *logger.Logger
}

// NotFound returns a no content response.
func (ncr *NoContentResultProvider) NotFound() Result {
	return noContent
}

// NotAuthorized returns a no content response.
func (ncr *NoContentResultProvider) NotAuthorized() Result {
	return noContent
}

// InternalError returns a no content response.
func (ncr *NoContentResultProvider) InternalError(err error) Result {
	if ncr.log != nil {
		ncr.log.Fatal(err)
	}
	return noContent
}

// BadRequest returns a no content response.
func (ncr *NoContentResultProvider) BadRequest(err error) Result {
	return noContent
}

// Result returns a no content response.
func (ncr *NoContentResultProvider) Result(response interface{}) Result {
	return noContent
}
