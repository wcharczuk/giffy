package web

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/julienschmidt/httprouter"
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
func (hc *HTTPContext) Param(name string) string {
	queryValue := hc.Request.URL.Query().Get(name)
	if len(queryValue) != 0 {
		return queryValue
	}

	headerValue := hc.Request.Header.Get(name)
	if len(headerValue) != 0 {
		return headerValue
	}

	formValue := hc.Request.FormValue(name)
	if len(formValue) != 0 {
		return formValue
	}

	cookie, cookieErr := hc.Request.Cookie(name)
	if cookieErr == nil && len(cookie.Value) != 0 {
		return cookie.Value
	}

	return util.StringEmpty
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
func (hc *HTTPContext) LogRequest() {
	Log(escapeRequestLogOutput(DefaultRequestLogFormat, hc))
}

// LogRequestWithW3CFormat consumes the context and writes a log message for the request.
func (hc *HTTPContext) LogRequestWithW3CFormat(format string) {
	Log(escapeRequestLogOutput(format, hc))
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
	c.Domain = hc.Request.Host
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

func (hc *HTTPContext) Raw(contentType string, body []byte) *RawResult {
	return &RawResult{ContentType: contentType, Body: body}
}

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

// --------------------------------------------------------------------------------
// Testing Methods & Types
// --------------------------------------------------------------------------------

type mockResponseWriter struct {
	contents   io.Writer
	statusCode int
	headers    http.Header
}

func (res *mockResponseWriter) Write(buffer []byte) (int, error) {
	return res.contents.Write(buffer)
}

func (res *mockResponseWriter) Header() http.Header {
	return res.headers
}

func (res *mockResponseWriter) WriteHeader(statusCode int) {
	res.statusCode = statusCode
}

// MockRequest returns a mock request.
func MockRequest(verb string, header http.Header, queryString url.Values, postBody *bytes.Buffer) *http.Request {
	url, _ := url.Parse("http://localhost/unit/test")
	r := http.Request{
		Method:     strings.ToUpper(verb),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Close:      true,
		Host:       "localhost",
		RemoteAddr: "127.0.0.1",
		RequestURI: "http://localhost/unit/test",
		URL:        url,
	}

	if postBody != nil && postBody.Len() != 0 {
		r.Body = ioutil.NopCloser(postBody)
	}

	if len(header) > 0 {
		for key, values := range header {
			for _, value := range values {
				r.Header.Set(key, value)
			}
		}
	}

	if len(queryString) > 0 {
		v := r.URL.Query()
		for key, values := range queryString {
			for _, value := range values {
				v.Add(key, value)
			}
		}
		r.URL.RawQuery = v.Encode()
		r.RequestURI = r.URL.String()
	}

	return &r
}

// MockResponse returns a mock response.
func MockResponse(responseBuffer *bytes.Buffer) http.ResponseWriter {
	return &mockResponseWriter{statusCode: http.StatusOK, contents: responseBuffer, headers: http.Header{}}
}

// MockHTTPContext returns a mocked HTTPContext.
func MockHTTPContext(verb string, params httprouter.Params, header http.Header, queryString url.Values, postBody *bytes.Buffer, responseBuffer *bytes.Buffer) *HTTPContext {
	mockRequest := MockRequest(verb, header, queryString, postBody)
	mockResponse := MockResponse(responseBuffer)

	if params == nil {
		params = httprouter.Params{}
	}

	return NewHTTPContext(mockResponse, mockRequest, params)
}
