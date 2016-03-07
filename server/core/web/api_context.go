package web

import (
	"net/http"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/julienschmidt/httprouter"
	"github.com/wcharczuk/giffy/server/model"
)

func NewAPIContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *APIContext {
	return &APIContext{
		Request:         r,
		Response:        w,
		RouteParameters: p,
	}
}

type APIContext struct {
	Response        http.ResponseWriter
	Request         *http.Request
	RouteParameters httprouter.Params
	User            *model.User

	statusCode    int
	requestStart  time.Time
	requestEnd    time.Time
	contentLength int
}

func (api APIContext) StatusCode() int {
	return api.statusCode
}

func (api *APIContext) SetStatusCode(code int) {
	api.statusCode = code
}

func (api APIContext) ContentLength() int {
	return api.contentLength
}

func (api *APIContext) SetContentLength(length int) {
	api.contentLength = length
}

func (api *APIContext) MarkRequestStart() {
	api.requestStart = time.Now().UTC()
}

func (api *APIContext) MarkRequestEnd() {
	api.requestEnd = time.Now().UTC()
}

func (api *APIContext) Elapsed() time.Duration {
	return api.requestEnd.Sub(api.requestStart)
}

func (api *APIContext) NotFound() *ServiceResponse {
	api.SetStatusCode(http.StatusNotFound)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HttpCode: http.StatusNotFound, Message: "Not Found."},
	}
}

func (api *APIContext) NoContent() *ServiceResponse {
	api.SetStatusCode(http.StatusNoContent)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HttpCode: http.StatusNoContent},
	}
}

func (api *APIContext) OK() *ServiceResponse {
	api.SetStatusCode(http.StatusOK)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HttpCode: http.StatusOK, Message: "OK!"},
	}
}

func (api *APIContext) InternalException(ex *exception.Exception) *ServiceResponse {
	api.SetStatusCode(http.StatusInternalServerError)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HttpCode: http.StatusInternalServerError, Message: "An internal server error occurred.", Exception: ex},
	}
}

func (api *APIContext) BadRequest(message string) *ServiceResponse {
	api.SetStatusCode(http.StatusBadRequest)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HttpCode: http.StatusBadRequest, Message: message},
	}
}

func (api *APIContext) Content(response interface{}) *ServiceResponse {
	api.SetStatusCode(http.StatusOK)
	return &ServiceResponse{
		Meta:     ServiceResponseMeta{HttpCode: http.StatusOK, Message: "OK!"},
		Response: response,
	}
}
