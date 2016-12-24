package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewAPIResultProvider Creates a new JSONResults object.
func NewAPIResultProvider(diag *logger.DiagnosticsAgent, r *RequestContext) *APIResultProvider {
	return &APIResultProvider{diagnostics: diag, requestContext: r}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	diagnostics    *logger.DiagnosticsAgent
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
	if ar.diagnostics != nil {
		if ar.requestContext != nil {
			ar.diagnostics.FatalWithReq(err, ar.requestContext.Request)
		} else {
			ar.diagnostics.FatalWithReq(err, nil)
		}
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
