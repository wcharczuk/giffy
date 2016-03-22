package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/core"
)

const (
	// PostBodySize is the maximum post body size we will typically consume.
	PostBodySize = int64(1 << 26) //64mb

	// PostBodySizeMax is the absolute maximum file size the server can handle.
	PostBodySizeMax = int64(1 << 32)
)

// PostedFile is a file that has been posted to an hc endpoint.
type PostedFile struct {
	Key      string
	Filename string
	Contents []byte
}

// NewHTTPContext returns a new hc context.
func NewHTTPContext(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HTTPContext {
	ctx := &HTTPContext{
		Request:         r,
		Response:        w,
		routeParameters: p,
		state:           map[string]interface{}{},
	}

	ctx.API = ProviderAPI
	ctx.View = ProviderView

	return ctx
}

// HTTPContext is the struct that represents the context for an hc request.
type HTTPContext struct {
	Response http.ResponseWriter
	Request  *http.Request

	API  *APIResultProvider
	View *ViewResultProvider

	state           map[string]interface{}
	routeParameters httprouter.Params

	statusCode    int
	contentLength int
	requestStart  time.Time
	requestEnd    time.Time
}

// State returns an object in the state cache.
func (hc *HTTPContext) State(key string) interface{} {
	if item, hasItem := hc.state[key]; hasItem {
		return item
	}
	return nil
}

// SetState sets the state for a key to an object.
func (hc *HTTPContext) SetState(key string, value interface{}) {
	hc.state[key] = value
}

// Param returns a parameter from the request.
func (hc *HTTPContext) Param(paramName string) string {
	return util.GetParamByName(hc.Request, paramName)
}

// PostBodyAsString is the string post body.
func (hc *HTTPContext) PostBodyAsString() string {
	defer hc.Request.Body.Close()
	bytes, _ := ioutil.ReadAll(hc.Request.Body)
	return string(bytes)
}

// PostBodyAsJSON reads the incoming post body (closing it) and marshals it to the target object as json.
func (hc *HTTPContext) PostBodyAsJSON(response interface{}) error {
	return DeserializeReaderAsJSON(response, hc.Request.Body)
}

// PostedFiles returns any files posted
func (hc *HTTPContext) PostedFiles() ([]PostedFile, error) {
	var files []PostedFile

	err := hc.Request.ParseMultipartForm(PostBodySize)
	if err == nil {
		for key := range hc.Request.MultipartForm.File {
			fileReader, fileHeader, err := hc.Request.FormFile(key)
			if err != nil {
				return nil, err
			}
			bytes, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}
			files = append(files, PostedFile{Key: key, Filename: fileHeader.Filename, Contents: bytes})
		}
	} else {
		err = hc.Request.ParseForm()
		if err == nil {
			for key := range hc.Request.PostForm {
				if fileReader, fileHeader, err := hc.Request.FormFile(key); err == nil && fileReader != nil {
					bytes, err := ioutil.ReadAll(fileReader)
					if err != nil {
						return nil, err
					}
					files = append(files, PostedFile{Key: key, Filename: fileHeader.Filename, Contents: bytes})
				}
			}
		}
	}
	return files, nil
}

// LogRequest consumes the context and writes a log message for the request.
func (hc *HTTPContext) LogRequest(format string) {
	fmt.Println(escapeRequestLogOutput(format, hc))
}

// RouteParameterInt returns a route parameter as an integer
func (hc *HTTPContext) RouteParameterInt(key string) int {
	v := hc.routeParameters.ByName(key)
	if !util.IsEmpty(v) {
		return util.ParseInt(v)
	}
	return int(0)
}

// RouteParameterInt64 returns a route parameter as an integer
func (hc *HTTPContext) RouteParameterInt64(key string) int64 {
	v := hc.routeParameters.ByName(key)
	if !util.IsEmpty(v) {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return vi
		}
	}
	return int64(0)
}

// RouteParameter returns a string route parameter
func (hc *HTTPContext) RouteParameter(key string) string {
	return hc.routeParameters.ByName(key)
}

// GetCookie returns a named cookie from the request.
func (hc *HTTPContext) GetCookie(name string) *http.Cookie {
	cookie, err := hc.Request.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}

// WriteCookie writes the cookie to the response.
func (hc *HTTPContext) WriteCookie(cookie *http.Cookie) {
	http.SetCookie(hc.Response, cookie)
}

// SetCookie is a helper method for WriteCookie.
func (hc *HTTPContext) SetCookie(name string, value string, expires *time.Time, path string) {
	c := http.Cookie{}
	c.Name = name
	c.HttpOnly = true
	c.Domain = core.ConfigHostname()
	c.Value = value
	c.Path = path
	if expires != nil {
		c.Expires = *expires
	}
	hc.WriteCookie(&c)
}

// ExtendCookieByDuration extends a cookie by a time duration (on the order of nanoseconds to hours).
func (hc *HTTPContext) ExtendCookieByDuration(name string, duration time.Duration) {
	cookie := hc.GetCookie(name)
	cookie.Expires = cookie.Expires.Add(duration)
	hc.WriteCookie(cookie)
}

// ExtendCookie extends a cookie by years, months or days.
func (hc HTTPContext) ExtendCookie(name string, years, months, days int) {
	cookie := hc.GetCookie(name)
	cookie.Expires.AddDate(years, months, days)
	hc.WriteCookie(cookie)
}

// ExpireCookie expires a cookie.
func (hc *HTTPContext) ExpireCookie(name string) {
	c := http.Cookie{}
	c.Name = name
	c.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	hc.WriteCookie(&c)
}

// Render writes the body of the response, it should not alter metadata.
func (hc *HTTPContext) Render(result ControllerResult) {
	renderErr := result.Render(hc)
	if renderErr != nil {
		LogError(renderErr)
	}
}

// --------------------------------------------------------------------------------
// Basic result providers
// --------------------------------------------------------------------------------

// NoContent returns a service response.
func (hc *HTTPContext) NoContent() *NoContentResult {
	return &NoContentResult{}
}

// Static returns a static result.
func (hc *HTTPContext) Static(filePath string) *StaticResult {
	return &StaticResult{
		FilePath: filePath,
	}
}

// Redirect returns a redirect result.
func (hc *HTTPContext) Redirect(path string) *RedirectResult {
	return &RedirectResult{
		RedirectURI: path,
	}
}

// --------------------------------------------------------------------------------
// Stats Methods used for logging.
// --------------------------------------------------------------------------------

// StatusCode returns the status code for the request, this is used for logging.
func (hc HTTPContext) getStatusCode() int {
	return hc.statusCode
}

// SetStatusCode sets the status code for the request, this is used for logging.
func (hc *HTTPContext) setStatusCode(code int) {
	hc.statusCode = code
}

// ContentLength returns the content length for the request, this is used for logging.
func (hc HTTPContext) getContentLength() int {
	return hc.contentLength
}

// SetContentLength sets the content length, this is used for logging.
func (hc *HTTPContext) setContentLength(length int) {
	hc.contentLength = length
}

// OnRequestStart will mark the start of request timing.
func (hc *HTTPContext) onRequestStart() {
	hc.requestStart = time.Now().UTC()
}

// OnRequestEnd will mark the end of request timing.
func (hc *HTTPContext) onRequestEnd() {
	hc.requestEnd = time.Now().UTC()
}

// Elapsed is the time delta between start and end.
func (hc *HTTPContext) elapsed() time.Duration {
	return hc.requestEnd.Sub(hc.requestStart)
}
