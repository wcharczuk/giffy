package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
)

// NewAPIResultProvider Creates a new JSONResults object.
func NewAPIResultProvider(app *App, r *RequestContext) *APIResultProvider {
	return &APIResultProvider{app: app, requestContext: r}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	app            *App
	requestContext *RequestContext
}

// NotFound returns a service response.
func (ar *APIResultProvider) NotFound() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusNotFound,
				Message:  "Not Found.",
			},
		},
	}
}

// NotAuthorized returns a service response.
func (ar *APIResultProvider) NotAuthorized() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusForbidden,
				Message:  "Not Authorized",
			},
		},
	}
}

// InternalError returns a service response.
func (ar *APIResultProvider) InternalError(err error) ControllerResult {
	if ar.app != nil {
		ar.app.onRequestError(ar.requestContext, err)
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response: &APIResponse{
				Meta: &APIResponseMeta{
					HTTPCode:  http.StatusInternalServerError,
					Message:   "An internal server error occurred.",
					Exception: exPtr,
				},
			},
		}
	}
	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusInternalServerError,
				Message:  err.Error(),
			},
		},
	}
}

// BadRequest returns a service response.
func (ar *APIResultProvider) BadRequest(message string) ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusBadRequest,
				Message:  message,
			},
		},
	}
}

// OK returns a service response.
func (ar *APIResultProvider) OK() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusOK,
				Message:  "OK!",
			},
		},
	}
}

// JSON returns a service response.
func (ar *APIResultProvider) JSON(response interface{}) ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				HTTPCode: http.StatusOK,
				Message:  "OK!",
			},
			Response: response,
		},
	}
}
