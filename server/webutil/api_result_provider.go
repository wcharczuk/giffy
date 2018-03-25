package webutil

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	web "github.com/blendlabs/go-web"
)

// API returns the api result provider.
func API(ctx *web.Ctx) *APIResultProvider {
	if typed, isTyped := ctx.DefaultResultProvider().(*APIResultProvider); isTyped {
		return typed
	}
	return NewAPIResultProvider(ctx)
}

// APIProviderAsDefault sets the context.CurrrentProvider() equal to context.API().
func APIProviderAsDefault(action web.Action) web.Action {
	return func(ctx *web.Ctx) web.Result {
		ctx.WithDefaultResultProvider(NewAPIResultProvider(ctx))
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
	return &APIResultProvider{log: r.Logger(), requestContext: r}
}

// APIResultProvider are context results for api methods.
type APIResultProvider struct {
	log            *logger.Logger
	requestContext *web.Ctx
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
	if ar.log != nil {
		if ar.requestContext != nil {
			ar.log.FatalWithReq(err, ar.requestContext.Request())
		} else {
			ar.log.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Ex); isException {
		return &web.JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response: &APIResponse{
				Meta: &APIResponseMeta{
					StatusCode: http.StatusInternalServerError,
					Message:    exPtr.Message(),
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
