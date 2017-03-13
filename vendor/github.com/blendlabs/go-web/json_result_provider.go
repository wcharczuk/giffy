package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewJSONResultProvider Creates a new JSONResults object.
func NewJSONResultProvider(diag *logger.Agent, r *Ctx) *JSONResultProvider {
	return &JSONResultProvider{diagnostics: diag, ctx: r}
}

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct {
	diagnostics *logger.Agent
	ctx         *Ctx
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
	if jrp.diagnostics != nil {
		if jrp.ctx != nil {
			jrp.diagnostics.FatalWithReq(err, jrp.ctx.Request)
		} else {
			jrp.diagnostics.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response:   exPtr,
		}
	}

	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (jrp *JSONResultProvider) BadRequest(message string) Result {
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response:   message,
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
