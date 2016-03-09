package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
)

const (
	// PostBodySize is the maximum post body size we will typically consume.
	PostBodySize = int64(1 << 26) //64mb

	// PostBodySizeMax is the absolute maximum file size the server can handle.
	PostBodySizeMax = int64(1 << 32)
)

// PostedFile is a file that has been posted to an API endpoint.
type PostedFile struct {
	Key      string
	Contents []byte
}

// NewAPIContext returns a new api context.
func NewAPIContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *APIContext {
	return &APIContext{
		Request:         r,
		Response:        w,
		routeParameters: p,
		state:           map[string]interface{}{},
	}
}

// APIContext is the struct that represents the context for an api request.
type APIContext struct {
	Response http.ResponseWriter
	Request  *http.Request

	state           map[string]interface{}
	routeParameters httprouter.Params

	statusCode    int
	requestStart  time.Time
	requestEnd    time.Time
	contentLength int
}

// State returns an object in the state cache.
func (api *APIContext) State(key string) interface{} {
	if item, hasItem := api.state[key]; hasItem {
		return item
	}
	return nil
}

// SetState sets the state for a key to an object.
func (api *APIContext) SetState(key string, value interface{}) {
	api.state[key] = value
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

// Param returns a parameter from the request.
func (api *APIContext) Param(paramName string) string {
	return util.GetParamByName(api.Request, paramName)
}

// PostBodyAsJSON reads the incoming post body (closing it) and marshals it to the target object as json.
func (api *APIContext) PostBodyAsJSON(response interface{}) error {
	return DeserializeReaderAsJSON(response, api.Request.Body)
}

// PostedFiles returns any files posted
func (api *APIContext) PostedFiles() ([]PostedFile, error) {
	err := api.Request.ParseMultipartForm(PostBodySize)
	if err != nil {
		return nil, err
	}

	var files []PostedFile
	for key := range api.Request.MultipartForm.File {
		fileReader, _, err := api.Request.FormFile(key)
		if err != nil {
			return nil, err
		}
		bytes, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return nil, err
		}
		files = append(files, PostedFile{Key: key, Contents: bytes})
	}
	return files, nil
}

// NotFound returns a service response.
func (api *APIContext) NotFound() *ServiceResponse {
	api.SetStatusCode(http.StatusNotFound)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusNotFound, Message: "Not Found."},
	}
}

// NotAuthorized returns a service response.
func (api *APIContext) NotAuthorized() *ServiceResponse {
	api.SetStatusCode(http.StatusForbidden)
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusForbidden, Message: "Not Authorized."},
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

// InternalError returns a service response.
func (api *APIContext) InternalError(err error) *ServiceResponse {
	api.SetStatusCode(http.StatusInternalServerError)
	if ex, isException := err.(*exception.Exception); isException {
		return &ServiceResponse{
			Meta: ServiceResponseMeta{HTTPCode: http.StatusInternalServerError, Message: "An internal server error occurred.", Exception: ex},
		}
	}
	return &ServiceResponse{
		Meta: ServiceResponseMeta{HTTPCode: http.StatusInternalServerError, Message: err.Error()},
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

// RouteParameterInt returns a route parameter as an integer
func (api *APIContext) RouteParameterInt(key string) int {
	v := api.routeParameters.ByName(key)
	if !util.IsEmpty(v) {
		return util.ParseInt(v)
	}
	return int(0)
}

// RouteParameterInt64 returns a route parameter as an integer
func (api *APIContext) RouteParameterInt64(key string) int64 {
	v := api.routeParameters.ByName(key)
	if !util.IsEmpty(v) {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return vi
		}
	}
	return int64(0)
}

// RouteParameter returns a string route parameter
func (api *APIContext) RouteParameter(key string) string {
	return api.routeParameters.ByName(key)
}
