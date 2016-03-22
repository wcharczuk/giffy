package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
)

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

// --------------------------------------------------------------------------------
// API result methods
// --------------------------------------------------------------------------------

// NewAPIResultProvider Creates a new APIResults object.
func NewAPIResultProvider() *APIResultProvider {
	return &APIResultProvider{}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct{}

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

// --------------------------------------------------------------------------------
// View result methods
// --------------------------------------------------------------------------------

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider() *ViewResultProvider {
	return &ViewResultProvider{}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct{}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
	}
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   "bad_request",
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   "error",
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   "not_found",
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusUnauthorized,
		ViewModel:  nil,
		Template:   "not_authorized",
	}
}
