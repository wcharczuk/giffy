package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewXMLResultProvider Creates a new JSONResults object.
func NewXMLResultProvider(diag *logger.Agent, ctx *Ctx) *XMLResultProvider {
	return &XMLResultProvider{diagnostics: diag, ctx: ctx}
}

// XMLResultProvider are context results for api methods.
type XMLResultProvider struct {
	diagnostics *logger.Agent
	ctx         *Ctx
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
	if xrp.diagnostics != nil {
		if xrp.ctx != nil {
			xrp.diagnostics.FatalWithReq(err, xrp.ctx.Request)
		} else {
			xrp.diagnostics.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &XMLResult{
			StatusCode: http.StatusInternalServerError,
			Response:   exPtr,
		}
	}

	return &XMLResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (xrp *XMLResultProvider) BadRequest(message string) Result {
	return &XMLResult{
		StatusCode: http.StatusBadRequest,
		Response:   message,
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
