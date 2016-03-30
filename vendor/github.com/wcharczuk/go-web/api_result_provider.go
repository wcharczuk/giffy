package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
)

// NewAPIResultProvider Creates a new APIResults object.
func NewAPIResultProvider(logger Logger) *APIResultProvider {
	return &APIResultProvider{logger: logger}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	logger Logger
}

// NotFound returns a service response.
func (ar *APIResultProvider) NotFound() ControllerResult {
	return &APIResult{
		Meta: &APIResultMeta{HTTPCode: http.StatusNotFound, Message: "Not Found."},
	}
}

// NotAuthorized returns a service response.
func (ar *APIResultProvider) NotAuthorized() ControllerResult {
	return &APIResult{
		Meta: &APIResultMeta{HTTPCode: http.StatusForbidden, Message: "Not Authorized."},
	}
}

// InternalError returns a service response.
func (ar *APIResultProvider) InternalError(err error) ControllerResult {
	ar.logger.Errorf("%v", err)
	if exPtr, isException := err.(*exception.Exception); isException {
		return &APIResult{
			Meta: &APIResultMeta{HTTPCode: http.StatusInternalServerError, Message: "An internal server error occurred.", Exception: exPtr},
		}
	}
	return &APIResult{
		Meta: &APIResultMeta{HTTPCode: http.StatusInternalServerError, Message: err.Error()},
	}
}

// BadRequest returns a service response.
func (ar *APIResultProvider) BadRequest(message string) ControllerResult {
	return &APIResult{
		Meta: &APIResultMeta{HTTPCode: http.StatusBadRequest, Message: message},
	}
}

// OK returns a service response.
func (ar *APIResultProvider) OK() ControllerResult {
	return &APIResult{
		Meta: &APIResultMeta{HTTPCode: http.StatusOK, Message: "OK!"},
	}
}

// JSON returns a service response.
func (ar *APIResultProvider) JSON(response interface{}) ControllerResult {
	return &APIResult{
		Meta:     &APIResultMeta{HTTPCode: http.StatusOK, Message: "OK!"},
		Response: response,
	}
}
