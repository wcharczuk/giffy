package web

import (
	"net/http"

	"github.com/blend/go-sdk/logger"
)

// NewJSONResultProvider Creates a new JSONResults object.
func NewJSONResultProvider(log *logger.Logger) *JSONResultProvider {
	return &JSONResultProvider{log: log}
}

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct {
	log *logger.Logger
}

// NotFound returns a service response.
func (jrp *JSONResultProvider) NotFound() Result {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (jrp *JSONResultProvider) NotAuthorized() Result {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (jrp *JSONResultProvider) InternalError(err error) Result {
	if jrp.log != nil {
		jrp.log.Fatal(err)
	}

	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (jrp *JSONResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &JSONResult{
			StatusCode: http.StatusBadRequest,
			Response:   err,
		}
	}
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response:   "Bad Request",
	}
}

// OK returns a service response.
func (jrp *JSONResultProvider) OK() Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns a json response.
func (jrp *JSONResultProvider) Result(response interface{}) Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   response,
	}
}
