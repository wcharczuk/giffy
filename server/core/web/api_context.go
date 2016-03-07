package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/model"
)

// NewAPIContext returns a new api context.
func NewAPIContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *APIContext {
	return &APIContext{
		Request:         r,
		Response:        w,
		RouteParameters: p,
	}
}

// APIContext is the struct that represents the context for an api request.
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

// StatusCode returns the status code for the request.
func (api APIContext) StatusCode() int {
	return api.statusCode
}

// SetStatusCode sets the status code for the request.
func (api *APIContext) SetStatusCode(code int) {
	api.statusCode = code
}

// ContentLength returns the content length for the request.
func (api APIContext) ContentLength() int {
	return api.contentLength
}

// SetContentLength sets the content length.
func (api *APIContext) SetContentLength(length int) {
	api.contentLength = length
}

// MarkRequestStart will mark the start of request timing.
func (api *APIContext) MarkRequestStart() {
	api.requestStart = time.Now().UTC()
}

// MarkRequestEnd will mark the end of request timing.
func (api *APIContext) MarkRequestEnd() {
	api.requestEnd = time.Now().UTC()
}

// Elapsed is the time delta between start and end.
func (api *APIContext) Elapsed() time.Duration {
	return api.requestEnd.Sub(api.requestStart)
}

// NotFound returns a service response.
func (api *APIContext) NotFound() *ServiceResponse {
	api.SetStatusCode(http.StatusNotFound)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusNotFound, Message: "Not Found."},
	}
}

// NoContent returns a service response.
func (api *APIContext) NoContent() *ServiceResponse {
	api.SetStatusCode(http.StatusNoContent)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusNoContent},
	}
}

// OK returns a service response.
func (api *APIContext) OK() *ServiceResponse {
	api.SetStatusCode(http.StatusOK)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusOK, Message: "OK!"},
	}
}

// InternalException returns a service response.
func (api *APIContext) InternalException(ex *exception.Exception) *ServiceResponse {
	api.SetStatusCode(http.StatusInternalServerError)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusInternalServerError, Message: "An internal server error occurred.", Exception: ex},
	}
}

// BadRequest returns a service response.
func (api *APIContext) BadRequest(message string) *ServiceResponse {
	api.SetStatusCode(http.StatusBadRequest)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusBadRequest, Message: message},
	}
}

// Content returns a service response.
func (api *APIContext) Content(response interface{}) *ServiceResponse {
	api.SetStatusCode(http.StatusOK)
	return &ServiceResponse{
		Meta:     ServiceResponseMeta{HTTPCode: http.StatusOK, Message: "OK!"},
		Response: response,
	}
}

// LogRequest consumes the context and writes a log message for the request.
func (api *APIContext) LogRequest(format string) {
	fmt.Println(escapeRequestLogOutput(format, api))
}
