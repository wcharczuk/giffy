package controller

import (
	"net/http"

	"github.com/blend/go-sdk/ex"
	exception "github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
)

// API returns the api result provider.
func API(ctx *web.Ctx) *APIResultProvider {
	if typed, isTyped := ctx.DefaultProvider.(*APIResultProvider); isTyped {
		return typed
	}
	return NewAPIResultProvider(ctx)
}

// APIProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func APIProviderAsDefault(action web.Action) web.Action {
	return func(ctx *web.Ctx) web.Result {
		ctx.DefaultProvider = NewAPIResultProvider(ctx)
		return action(ctx)
	}
}

// APIResponseMeta is the meta component of a service response.
type APIResponseMeta struct {
	StatusCode int
	Message    string        `json:",omitempty"`
	Exception  *exception.Ex `json:",omitempty"`
}

// APIResponse is the standard API response format.
type APIResponse struct {
	Meta     *APIResponseMeta
	Response interface{}
}

// NewAPIResultProvider Creates a new JSONResults object.
func NewAPIResultProvider(r *web.Ctx) *APIResultProvider {
	return &APIResultProvider{log: r.App.Log, requestContext: r}
}

var (
	_ web.ResultProvider = (*APIResultProvider)(nil)
)

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	log            logger.Log
	requestContext *web.Ctx
}

// Status returns a service response.
func (ar *APIResultProvider) Status(statusCode int, data interface{}) web.Result {
	return &web.JSONResult{
		StatusCode: statusCode,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: statusCode,
			},
			Response: data,
		},
	}
}

// NotFound returns a service response.
func (ar *APIResultProvider) NotFound() web.Result {
	return &web.JSONResult{
		StatusCode: http.StatusNotFound,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusNotFound,
				Message:    "Not Found.",
			},
		},
	}
}

// NotAuthorized returns a service response.
func (ar *APIResultProvider) NotAuthorized() web.Result {
	return &web.JSONResult{
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
func (ar *APIResultProvider) InternalError(err error) web.Result {
	logger.MaybeFatal(ar.log, err)

	if exPtr, isException := err.(*exception.Ex); isException {
		return &web.JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response: &APIResponse{
				Meta: &APIResponseMeta{
					StatusCode: http.StatusInternalServerError,
					Message:    ex.ErrMessage(exPtr),
					Exception:  exPtr,
				},
			},
		}
	}
	return &web.JSONResult{
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
func (ar *APIResultProvider) BadRequest(err error) web.Result {
	return &web.JSONResult{
		StatusCode: http.StatusBadRequest,
		Response: &APIResponse{
			Meta: &APIResponseMeta{
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			},
		},
	}
}

// OK returns a service response.
func (ar *APIResultProvider) OK() web.Result {
	return &web.JSONResult{
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
func (ar *APIResultProvider) Result(response interface{}) web.Result {
	return &web.JSONResult{
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
