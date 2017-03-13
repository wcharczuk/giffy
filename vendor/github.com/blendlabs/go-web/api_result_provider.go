package web

import (
	"fmt"
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewAPIResultProvider Creates a new JSONResults object.
func NewAPIResultProvider(diag *logger.Agent, ctx *Ctx) *APIResultProvider {
	return &APIResultProvider{diagnostics: diag, ctx: ctx}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	diagnostics *logger.Agent
	ctx         *Ctx
}

// NotFound returns a service response.
func (ar *APIResultProvider) NotFound() Result {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusNotFound,
				Message:    "Not Found",
			},
		},
	}
}

// NotAuthorized returns a service response.
func (ar *APIResultProvider) NotAuthorized() Result {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusForbidden,
				Message:    "Not Authorized",
			},
		},
	}
}

// InternalError returns a service response.
func (ar *APIResultProvider) InternalError(err error) Result {
	if ar.diagnostics != nil {
		if ar.ctx != nil {
			ar.diagnostics.FatalWithReq(err, ar.ctx.Request)
		} else {
			ar.diagnostics.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response: &APIResponse{
				Meta: &APIResponseMeta{
					StatusCode: http.StatusInternalServerError,
					Message:    fmt.Sprintf("%v", exPtr),
					Exception:  exPtr,
				},
			},
		}
	}
	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			},
		},
	}
}

// BadRequest returns a service response.
func (ar *APIResultProvider) BadRequest(message string) Result {
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusBadRequest,
				Message:    message,
			},
		},
	}
}

// OK returns a service response.
func (ar *APIResultProvider) OK() Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusOK,
				Message:    "OK!",
			},
		},
	}
}

// Result returns a service response.
func (ar *APIResultProvider) Result(response interface{}) Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusOK,
				Message:    "OK!",
			},
			Response: response,
		},
	}
}
