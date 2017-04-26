package web

import "net/http"

// NewJSONResultProvider Creates a new JSONResults object.
func NewJSONResultProvider(ctx *Ctx) *JSONResultProvider {
	return &JSONResultProvider{ctx: ctx}
}

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct {
	ctx *Ctx
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
	if jrp.ctx != nil {
		jrp.ctx.logFatal(err)
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
