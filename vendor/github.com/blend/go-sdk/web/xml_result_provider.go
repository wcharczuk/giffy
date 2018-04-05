package web

import (
	"net/http"

	"github.com/blend/go-sdk/logger"
)

// NewXMLResultProvider Creates a new JSONResults object.
func NewXMLResultProvider(log *logger.Logger) *XMLResultProvider {
	return &XMLResultProvider{log: log}
}

// XMLResultProvider are context results for api methods.
type XMLResultProvider struct {
	log *logger.Logger
}

// NotFound returns a service response.
func (xrp *XMLResultProvider) NotFound() Result {
	return &XMLResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (xrp *XMLResultProvider) NotAuthorized() Result {
	return &XMLResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (xrp *XMLResultProvider) InternalError(err error) Result {
	if xrp.log != nil {
		xrp.log.Fatal(err)
	}

	return &XMLResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (xrp *XMLResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &XMLResult{
			StatusCode: http.StatusBadRequest,
			Response:   err,
		}
	}
	return &XMLResult{
		StatusCode: http.StatusBadRequest,
		Response:   "Bad Request",
	}
}

// OK returns a service response.
func (xrp *XMLResultProvider) OK() Result {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns an xml response.
func (xrp *XMLResultProvider) Result(response interface{}) Result {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   response,
	}
}
