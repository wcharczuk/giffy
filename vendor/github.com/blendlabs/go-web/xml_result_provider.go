package web

import "net/http"

// NewXMLResultProvider Creates a new JSONResults object.
func NewXMLResultProvider(ctx *Ctx) *XMLResultProvider {
	return &XMLResultProvider{ctx: ctx}
}

// XMLResultProvider are context results for api methods.
type XMLResultProvider struct {
	ctx *Ctx
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
	if xrp.ctx != nil {
		xrp.ctx.logFatal(err)
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
