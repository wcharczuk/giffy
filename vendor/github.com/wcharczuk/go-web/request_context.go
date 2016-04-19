package web

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/blendlabs/go-util"
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

// State is the collection of state objects on a context.
type State map[string]interface{}

// RequestEventHandler is an event handler for requests.
type RequestEventHandler func(r *RequestContext)

// RequestEventErrorHandler is fired when an error occurs.
type RequestEventErrorHandler func(r *RequestContext, err interface{})

// NewRequestContext returns a new hc context.
func NewRequestContext(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	ctx := &RequestContext{
		Response:        w,
		Request:         r,
		routeParameters: p,
		state:           State{},
	}

	return ctx
}

// RequestContext is the struct that represents the context for an hc request.
type RequestContext struct {
	Response ResponseWriter
	Request  *http.Request

	api             *APIResultProvider
	view            *ViewResultProvider
	currentProvider ControllerResultProvider

	logger Logger

	tx *sql.Tx

	state           State
	routeParameters RouteParameters

	statusCode    int
	contentLength int
	requestStart  time.Time
	requestEnd    time.Time
}

// isolateTo isolates a request context to a transaction.
func (rc *RequestContext) isolateTo(tx *sql.Tx) *RequestContext {
	rc.tx = tx
	return rc
}

// Tx returns the transaction a request context may or may not be isolated to.
func (rc *RequestContext) Tx() *sql.Tx {
	return rc.tx
}

// TxBegin either returns the existing (testing) transaction on the request, or calls the provider.
func (rc *RequestContext) TxBegin(provider func() (*sql.Tx, error)) (*sql.Tx, error) {
	if rc.tx != nil {
		return rc.tx, nil
	}
	return provider()
}

// TxCommit will call the committer if the request context is not isolated to a transaction already.
func (rc *RequestContext) TxCommit(commiter func() error) error {
	if rc.tx != nil {
		return nil
	}
	return commiter()
}

// TxRollback will call the rollbacker if the request context is not isolated to a transaction already.
func (rc *RequestContext) TxRollback(rollbacker func() error) error {
	if rc.tx != nil {
		return nil
	}
	return rollbacker()
}

// API returns the API result provider.
func (rc *RequestContext) API() *APIResultProvider {
	return rc.api
}

// View returns the view result provider.
func (rc *RequestContext) View() *ViewResultProvider {
	return rc.view
}

// CurrentProvider returns the current result provider for the context. This is
// set by calling SetCurrentProvider or using one of the pre-built middleware
// steps that set it for you.
func (rc *RequestContext) CurrentProvider() ControllerResultProvider {
	return rc.currentProvider
}

// SetCurrentProvider sets the current result provider.
func (rc *RequestContext) SetCurrentProvider(provider ControllerResultProvider) {
	rc.currentProvider = provider
}

// State returns an object in the state cache.
func (rc *RequestContext) State(key string) interface{} {
	if item, hasItem := rc.state[key]; hasItem {
		return item
	}
	return nil
}

// SetState sets the state for a key to an object.
func (rc *RequestContext) SetState(key string, value interface{}) {
	rc.state[key] = value
}

// Param returns a parameter from the request.
func (rc *RequestContext) Param(name string) string {
	queryValue := rc.Request.URL.Query().Get(name)
	if len(queryValue) != 0 {
		return queryValue
	}

	headerValue := rc.Request.Header.Get(name)
	if len(headerValue) != 0 {
		return headerValue
	}

	formValue := rc.Request.FormValue(name)
	if len(formValue) != 0 {
		return formValue
	}

	cookie, cookieErr := rc.Request.Cookie(name)
	if cookieErr == nil && len(cookie.Value) != 0 {
		return cookie.Value
	}

	return util.StringEmpty
}

// PostBodyAsString is the string post body.
func (rc *RequestContext) PostBodyAsString() string {
	defer rc.Request.Body.Close()
	bytes, _ := ioutil.ReadAll(rc.Request.Body)
	return string(bytes)
}

// PostBodyAsJSON reads the incoming post body (closing it) and marshals it to the target object as json.
func (rc *RequestContext) PostBodyAsJSON(response interface{}) error {
	return DeserializeReaderAsJSON(response, rc.Request.Body)
}

// PostedFiles returns any files posted
func (rc *RequestContext) PostedFiles() ([]PostedFile, error) {
	var files []PostedFile

	err := rc.Request.ParseMultipartForm(PostBodySize)
	if err == nil {
		for key := range rc.Request.MultipartForm.File {
			fileReader, fileHeader, err := rc.Request.FormFile(key)
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
		err = rc.Request.ParseForm()
		if err == nil {
			for key := range rc.Request.PostForm {
				if fileReader, fileHeader, err := rc.Request.FormFile(key); err == nil && fileReader != nil {
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

// RouteParameterInt returns a route parameter as an integer
func (rc *RequestContext) RouteParameterInt(key string) int {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return util.ParseInt(value)
	}
	return int(0)
}

// RouteParameterInt64 returns a route parameter as an integer
func (rc *RequestContext) RouteParameterInt64(key string) int64 {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		valueAsInt, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return valueAsInt
		}
	}
	return int64(0)
}

// RouteParameter returns a string route parameter
func (rc *RequestContext) RouteParameter(key string) string {
	return rc.routeParameters[key]
}

// GetCookie returns a named cookie from the request.
func (rc *RequestContext) GetCookie(name string) *http.Cookie {
	cookie, err := rc.Request.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}

// WriteCookie writes the cookie to the response.
func (rc *RequestContext) WriteCookie(cookie *http.Cookie) {
	http.SetCookie(rc.Response, cookie)
}

// SetCookie is a helper method for WriteCookie.
func (rc *RequestContext) SetCookie(name string, value string, expires *time.Time, path string) {
	c := http.Cookie{}
	c.Name = name
	c.HttpOnly = true
	c.Domain = rc.Request.Host
	c.Value = value
	c.Path = path
	if expires != nil {
		c.Expires = *expires
	}
	rc.WriteCookie(&c)
}

// ExtendCookieByDuration extends a cookie by a time duration (on the order of nanoseconds to hours).
func (rc *RequestContext) ExtendCookieByDuration(name string, duration time.Duration) {
	cookie := rc.GetCookie(name)
	cookie.Expires = cookie.Expires.Add(duration)
	rc.WriteCookie(cookie)
}

// ExtendCookie extends a cookie by years, months or days.
func (rc *RequestContext) ExtendCookie(name string, years, months, days int) {
	cookie := rc.GetCookie(name)
	cookie.Expires.AddDate(years, months, days)
	rc.WriteCookie(cookie)
}

// ExpireCookie expires a cookie.
func (rc *RequestContext) ExpireCookie(name string) {
	c := http.Cookie{}
	c.Name = name
	c.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	rc.WriteCookie(&c)
}

// --------------------------------------------------------------------------------
// Logging
// --------------------------------------------------------------------------------

// LogRequest consumes the context and writes a log message for the request.
func (rc *RequestContext) LogRequest() {
	if rc.logger != nil {
		rc.logger.Log(FormatRequestLog(DefaultRequestLogFormat, rc))
	}
}

// LogRequestWithW3CFormat consumes the context and writes a log message for the request.
func (rc *RequestContext) LogRequestWithW3CFormat(format string) {
	if rc.logger != nil {
		rc.logger.Log(FormatRequestLog(format, rc))
	}
}

// Logger returns the logger.
func (rc *RequestContext) Logger() Logger {
	return rc.logger
}

// --------------------------------------------------------------------------------
// Basic result providers
// --------------------------------------------------------------------------------

// Raw returns a binary response body, sniffing the content type.
func (rc *RequestContext) Raw(body []byte) *RawResult {
	sniffedContentType := http.DetectContentType(body)
	return rc.RawWithContentType(sniffedContentType, body)
}

// RawWithContentType returns a binary response with a given content type.
func (rc *RequestContext) RawWithContentType(contentType string, body []byte) *RawResult {
	return &RawResult{ContentType: contentType, Body: body}
}

// JSON returns a basic json result.
func (rc *RequestContext) JSON(object interface{}) *JSONResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   object,
	}
}

// NoContent returns a service response.
func (rc *RequestContext) NoContent() *NoContentResult {
	return &NoContentResult{}
}

// Static returns a static result.
func (rc *RequestContext) Static(filePath string) *StaticResult {
	return NewStaticResultForSingleFile(filePath)
}

// Redirect returns a redirect result.
func (rc *RequestContext) Redirect(path string) *RedirectResult {
	return &RedirectResult{
		RedirectURI: path,
	}
}

// --------------------------------------------------------------------------------
// Stats Methods used for logging.
// --------------------------------------------------------------------------------

// StatusCode returns the status code for the request, this is used for logging.
func (rc *RequestContext) getStatusCode() int {
	return rc.statusCode
}

// SetStatusCode sets the status code for the request, this is used for logging.
func (rc *RequestContext) setStatusCode(code int) {
	rc.statusCode = code
}

// ContentLength returns the content length for the request, this is used for logging.
func (rc *RequestContext) getContentLength() int {
	return rc.contentLength
}

// SetContentLength sets the content length, this is used for logging.
func (rc *RequestContext) setContentLength(length int) {
	rc.contentLength = length
}

// OnRequestStart will mark the start of request timing.
func (rc *RequestContext) onRequestStart() {
	rc.requestStart = time.Now().UTC()
}

// OnRequestEnd will mark the end of request timing.
func (rc *RequestContext) onRequestEnd() {
	rc.requestEnd = time.Now().UTC()
}

// Elapsed is the time delta between start and end.
func (rc *RequestContext) Elapsed() time.Duration {
	return rc.requestEnd.Sub(rc.requestStart)
}
