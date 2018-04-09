package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/blend/go-sdk/logger"
)

const (
	// PostBodySize is the maximum post body size we will typically consume.
	PostBodySize = int64(1 << 26) //64mb

	// PostBodySizeMax is the absolute maximum file size the server can handle.
	PostBodySizeMax = int64(1 << 32) //enormous.

	// StringEmpty is the empty string.
	StringEmpty = ""
)

// Request is an alias to Ctx.
// It is part of a longer term transition.
type Request = Ctx

// defaultResultProvider is used by bare ctx results, it generally
// won't stay the default for long, as it's overwritten by `App`.
var defaultResultProvider = &TextResultProvider{}

// NewCtx returns a new hc context.
func NewCtx(w ResponseWriter, r *http.Request, p RouteParameters, s State) *Ctx {
	ctx := &Ctx{
		response:        w,
		request:         r,
		routeParameters: p,
		state:           s,
		defaultResultProvider: defaultResultProvider,
	}

	if ctx.state == nil {
		ctx.state = State{}
	}

	return ctx
}

// Ctx is the struct that represents the context for an hc request.
type Ctx struct {
	response ResponseWriter
	request  *http.Request

	app  *App
	log  *logger.Logger
	auth *AuthManager

	postBody []byte

	view                  *ViewResultProvider
	json                  *JSONResultProvider
	xml                   *XMLResultProvider
	text                  *TextResultProvider
	defaultResultProvider ResultProvider

	state            State
	routeParameters  RouteParameters
	route            *Route
	statusCode       int
	contentLength    int
	requestStart     time.Time
	requestEnd       time.Time
	requestLogFormat string
	session          *Session

	ctx    context.Context
	cancel context.CancelFunc
}

// WithResponse sets the underlying response.
func (rc *Ctx) WithResponse(res ResponseWriter) *Ctx {
	rc.response = res
	return rc
}

// Response returns the underyling response.
func (rc *Ctx) Response() ResponseWriter {
	return rc.response
}

// WithRequest sets the underlying request.
func (rc *Ctx) WithRequest(req *http.Request) *Ctx {
	rc.request = req
	return rc
}

// Request returns the underlying request.
func (rc *Ctx) Request() *http.Request {
	return rc.request
}

// WithTx sets a transaction on the context.
func (rc *Ctx) WithTx(tx *sql.Tx, optionalKey ...string) *Ctx {
	WithTx(rc, tx, optionalKey...)
	return rc
}

// WithContext sets the background context for the request.
func (rc *Ctx) WithContext(ctx context.Context) *Ctx {
	rc.ctx = ctx
	return rc
}

// Background returns the background context for a request.
func (rc *Ctx) Background() context.Context {
	return rc.ctx
}

// Cancel calls the cancel func if it's set.
func (rc *Ctx) Cancel() {
	if rc.cancel != nil {
		rc.cancel()
	}
}

// WithApp sets the app reference for the ctx.
func (rc *Ctx) WithApp(app *App) *Ctx {
	rc.app = app
	return rc
}

// App returns the app reference.
func (rc *Ctx) App() *App {
	return rc.app
}

// WithAuth sets the request context auth.
func (rc *Ctx) WithAuth(authManager *AuthManager) *Ctx {
	rc.auth = authManager
	return rc
}

// Auth returns the AuthManager for the request.
func (rc *Ctx) Auth() *AuthManager {
	return rc.auth
}

// WithSession sets the session for the request.
func (rc *Ctx) WithSession(session *Session) *Ctx {
	rc.session = session
	return rc
}

// Session returns the session (if any) on the request.
func (rc *Ctx) Session() *Session {
	return rc.session
}

// View returns the view result provider.
func (rc *Ctx) View() *ViewResultProvider {
	return rc.view
}

// JSON returns the JSON result provider.
func (rc *Ctx) JSON() *JSONResultProvider {
	return rc.json
}

// XML returns the xml result provider.
func (rc *Ctx) XML() *XMLResultProvider {
	return rc.xml
}

// Text returns the text result provider.
func (rc *Ctx) Text() *TextResultProvider {
	return rc.text
}

// DefaultResultProvider returns the current result provider for the context. This is
// set by calling SetDefaultResultProvider or using one of the pre-built middleware
// steps that set it for you.
func (rc *Ctx) DefaultResultProvider() ResultProvider {
	return rc.defaultResultProvider
}

// WithDefaultResultProvider sets the default result provider.
func (rc *Ctx) WithDefaultResultProvider(provider ResultProvider) *Ctx {
	rc.defaultResultProvider = provider
	return rc
}

// State returns the full state bag.
func (rc *Ctx) State() State {
	return rc.state
}

// StateValue returns an object in the state cache.
func (rc *Ctx) StateValue(key string) interface{} {
	if rc.state == nil {
		return nil
	}
	if value, hasValue := rc.state[key]; hasValue {
		return value
	}
	return nil
}

// WithStateValue sets the state for a key to an object.
func (rc *Ctx) WithStateValue(key string, value interface{}) *Ctx {
	if rc.state == nil {
		rc.state = State{}
	}
	rc.state[key] = value
	return rc
}

// ParamString is a shortcut for ParamString that swallows the missing value error.
func (rc *Ctx) ParamString(name string) string {
	value, _ := rc.Param(name)
	return value
}

// Param returns a parameter from the request.
func (rc *Ctx) Param(name string) (string, error) {
	if rc.routeParameters != nil {
		routeValue := rc.routeParameters.Get(name)
		if len(routeValue) > 0 {
			return routeValue, nil
		}
	}
	if rc.request != nil {
		if rc.request.URL != nil {
			queryValue := rc.request.URL.Query().Get(name)
			if len(queryValue) > 0 {
				return queryValue, nil
			}
		}
		if rc.request.Header != nil {
			headerValue := rc.request.Header.Get(name)
			if len(headerValue) > 0 {
				return headerValue, nil
			}
		}

		formValue := rc.request.FormValue(name)
		if len(formValue) > 0 {
			return formValue, nil
		}

		cookie, cookieErr := rc.request.Cookie(name)
		if cookieErr == nil && len(cookie.Value) != 0 {
			return cookie.Value, nil
		}
	}

	return "", newParameterMissingError(name)
}

// ParamInt returns a parameter from any location as an integer.
func (rc *Ctx) ParamInt(name string) (int, error) {
	paramValue, err := rc.Param(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(paramValue)
}

// ParamInt64 returns a parameter from any location as an int64.
func (rc *Ctx) ParamInt64(name string) (int64, error) {
	paramValue, err := rc.Param(name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(paramValue, 10, 64)
}

// ParamFloat64 returns a parameter from any location as a float64.
func (rc *Ctx) ParamFloat64(name string) (float64, error) {
	paramValue, err := rc.Param(name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(paramValue, 64)
}

// ParamTime returns a parameter from any location as a time with a given format.
func (rc *Ctx) ParamTime(name, format string) (time.Time, error) {
	paramValue, err := rc.Param(name)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(format, paramValue)
}

// ParamBool returns a boolean value for a param.
func (rc *Ctx) ParamBool(name string) (bool, error) {
	paramValue, err := rc.Param(name)
	if err != nil {
		return false, err
	}
	lower := strings.ToLower(paramValue)
	return lower == "true" || lower == "1" || lower == "yes", nil
}

// PostBody returns the bytes in a post body.
func (rc *Ctx) PostBody() ([]byte, error) {
	var err error
	if len(rc.postBody) == 0 {
		if rc.request != nil && rc.request.Body != nil {
			defer rc.request.Body.Close()
			rc.postBody, err = ioutil.ReadAll(rc.request.Body)
		}
		if err != nil {
			return nil, err
		}
	}
	return rc.postBody, nil
}

// PostBodyAsString returns the post body as a string.
func (rc *Ctx) PostBodyAsString() (string, error) {
	body, err := rc.PostBody()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// PostBodyAsJSON reads the incoming post body (closing it) and marshals it to the target object as json.
func (rc *Ctx) PostBodyAsJSON(response interface{}) error {
	body, err := rc.PostBody()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, response)
}

// PostBodyAsXML reads the incoming post body (closing it) and marshals it to the target object as xml.
func (rc *Ctx) PostBodyAsXML(response interface{}) error {
	body, err := rc.PostBody()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, response)
}

// PostedFiles returns any files posted
func (rc *Ctx) PostedFiles() ([]PostedFile, error) {
	var files []PostedFile

	err := rc.request.ParseMultipartForm(PostBodySize)
	if err == nil {
		for key := range rc.request.MultipartForm.File {
			fileReader, fileHeader, err := rc.request.FormFile(key)
			if err != nil {
				return nil, err
			}
			bytes, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}
			files = append(files, PostedFile{Key: key, FileName: fileHeader.Filename, Contents: bytes})
		}
	} else {
		err = rc.request.ParseForm()
		if err == nil {
			for key := range rc.request.PostForm {
				if fileReader, fileHeader, err := rc.request.FormFile(key); err == nil && fileReader != nil {
					bytes, err := ioutil.ReadAll(fileReader)
					if err != nil {
						return nil, err
					}
					files = append(files, PostedFile{Key: key, FileName: fileHeader.Filename, Contents: bytes})
				}
			}
		}
	}
	return files, nil
}

func newParameterMissingError(paramName string) error {
	return fmt.Errorf("`%s` parameter is missing", paramName)
}

// RouteParamInt returns a route parameter as an integer.
func (rc *Ctx) RouteParamInt(key string) (int, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.Atoi(value)
	}
	return 0, newParameterMissingError(key)
}

// RouteParamInt64 returns a route parameter as an integer.
func (rc *Ctx) RouteParamInt64(key string) (int64, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, newParameterMissingError(key)
}

// RouteParamFloat64 returns a route parameter as an float64.
func (rc *Ctx) RouteParamFloat64(key string) (float64, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.ParseFloat(value, 64)
	}
	return 0, newParameterMissingError(key)
}

// RouteParam returns a string route parameter
func (rc *Ctx) RouteParam(key string) (string, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return value, nil
	}
	return StringEmpty, newParameterMissingError(key)
}

// QueryParam returns a query parameter.
func (rc *Ctx) QueryParam(key string) (string, error) {
	if value := rc.request.URL.Query().Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, newParameterMissingError(key)
}

// QueryParamInt returns a query parameter as an integer.
func (rc *Ctx) QueryParamInt(key string) (int, error) {
	if value := rc.request.URL.Query().Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamInt64 returns a query parameter as an int64.
func (rc *Ctx) QueryParamInt64(key string) (int64, error) {
	if value := rc.request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamFloat64 returns a query parameter as a float64.
func (rc *Ctx) QueryParamFloat64(key string) (float64, error) {
	if value := rc.request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamTime returns a query parameter as a time.Time.
func (rc *Ctx) QueryParamTime(key, format string) (time.Time, error) {
	if value := rc.request.URL.Query().Get(key); len(value) > 0 {
		return time.Parse(format, value)
	}
	return time.Time{}, newParameterMissingError(key)
}

// HeaderParam returns a header parameter value.
func (rc *Ctx) HeaderParam(key string) (string, error) {
	if value := rc.request.Header.Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, newParameterMissingError(key)
}

// HeaderParamInt returns a header parameter value as an integer.
func (rc *Ctx) HeaderParamInt(key string) (int, error) {
	if value := rc.request.Header.Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamInt64 returns a header parameter value as an integer.
func (rc *Ctx) HeaderParamInt64(key string) (int64, error) {
	if value := rc.request.Header.Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamFloat64 returns a header parameter value as an float64.
func (rc *Ctx) HeaderParamFloat64(key string) (float64, error) {
	if value := rc.request.Header.Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamTime returns a header parameter value as an float64.
func (rc *Ctx) HeaderParamTime(key, format string) (time.Time, error) {
	if value := rc.request.Header.Get(key); len(value) > 0 {
		return time.Parse(format, key)
	}
	return time.Time{}, newParameterMissingError(key)
}

// GetCookie returns a named cookie from the request.
func (rc *Ctx) GetCookie(name string) *http.Cookie {
	cookie, err := rc.request.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}

// WriteCookie writes the cookie to the response.
func (rc *Ctx) WriteCookie(cookie *http.Cookie) {
	http.SetCookie(rc.response, cookie)
}

func (rc *Ctx) getCookieDomain() string {
	if rc.app != nil && rc.app.baseURL != nil {
		return rc.app.baseURL.Host
	}
	return rc.request.Host
}

// WriteNewCookie is a helper method for WriteCookie.
func (rc *Ctx) WriteNewCookie(name string, value string, expires *time.Time, path string, secure bool) {
	c := http.Cookie{
		Name:     name,
		HttpOnly: true,
		Value:    value,
		Path:     path,
		Secure:   secure,
		Domain:   rc.getCookieDomain(),
	}
	if expires != nil {
		c.Expires = *expires
	}
	rc.WriteCookie(&c)
}

// ExtendCookieByDuration extends a cookie by a time duration (on the order of nanoseconds to hours).
func (rc *Ctx) ExtendCookieByDuration(name string, path string, duration time.Duration) {
	c := rc.GetCookie(name)
	if c == nil {
		return
	}
	c.Path = path
	c.Domain = rc.getCookieDomain()
	c.Expires = c.Expires.Add(duration)
	rc.WriteCookie(c)
}

// ExtendCookie extends a cookie by years, months or days.
func (rc *Ctx) ExtendCookie(name string, path string, years, months, days int) {
	c := rc.GetCookie(name)
	if c == nil {
		return
	}
	c.Path = path
	c.Domain = rc.getCookieDomain()
	c.Expires.AddDate(years, months, days)
	rc.WriteCookie(c)
}

// ExpireCookie expires a cookie.
func (rc *Ctx) ExpireCookie(name string, path string) {
	c := rc.GetCookie(name)
	if c == nil {
		return
	}
	c.Path = path
	c.Value = NewSessionID()
	c.Domain = rc.getCookieDomain()
	c.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	rc.WriteCookie(c)
}

// --------------------------------------------------------------------------------
// Diagnostics
// --------------------------------------------------------------------------------

// Logger returns the diagnostics agent.
func (rc *Ctx) Logger() *logger.Logger {
	return rc.log
}

// --------------------------------------------------------------------------------
// Basic result providers
// --------------------------------------------------------------------------------

// Raw returns a binary response body, sniffing the content type.
func (rc *Ctx) Raw(body []byte) *RawResult {
	sniffedContentType := http.DetectContentType(body)
	return rc.RawWithContentType(sniffedContentType, body)
}

// RawWithContentType returns a binary response with a given content type.
func (rc *Ctx) RawWithContentType(contentType string, body []byte) *RawResult {
	return &RawResult{ContentType: contentType, Body: body}
}

// NoContent returns a service response.
func (rc *Ctx) NoContent() *NoContentResult {
	return &NoContentResult{}
}

// Static returns a static result.
func (rc *Ctx) Static(filePath string) *StaticResult {
	return NewStaticResultForFile(filePath)
}

// Redirectf returns a redirect result.
func (rc *Ctx) Redirectf(format string, args ...interface{}) *RedirectResult {
	if len(args) > 0 {
		return &RedirectResult{
			RedirectURI: fmt.Sprintf(format, args...),
		}
	}
	return &RedirectResult{
		RedirectURI: format,
	}
}

// RedirectWithMethodf returns a redirect result with a given method.
func (rc *Ctx) RedirectWithMethodf(method, format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		Method:      method,
		RedirectURI: fmt.Sprintf(format, args...),
	}
}

// --------------------------------------------------------------------------------
// Stats Methods used for logging.
// --------------------------------------------------------------------------------

// StatusCode returns the status code for the request, this is used for logging.
func (rc *Ctx) getLoggedStatusCode() int {
	return rc.statusCode
}

// SetStatusCode sets the status code for the request, this is used for logging.
func (rc *Ctx) setLoggedStatusCode(code int) {
	rc.statusCode = code
}

// ContentLength returns the content length for the request, this is used for logging.
func (rc *Ctx) getLoggedContentLength() int {
	return rc.contentLength
}

// SetContentLength sets the content length, this is used for logging.
func (rc *Ctx) setLoggedContentLength(length int) {
	rc.contentLength = length
}

// OnRequestStart will mark the start of request timing.
func (rc *Ctx) onRequestStart() {
	rc.requestStart = time.Now().UTC()
}

// Start returns the request start time.
func (rc Ctx) Start() time.Time {
	return rc.requestStart
}

// OnRequestEnd will mark the end of request timing.
func (rc *Ctx) onRequestEnd() {
	rc.requestEnd = time.Now().UTC()
}

// Elapsed is the time delta between start and end.
func (rc *Ctx) Elapsed() time.Duration {
	if !rc.requestEnd.IsZero() {
		return rc.requestEnd.Sub(rc.requestStart)
	}
	return time.Now().UTC().Sub(rc.requestStart)
}

// Route returns the original route match for the request.
func (rc *Ctx) Route() *Route {
	return rc.route
}

// PostedFile is a file that has been posted to an hc endpoint.
type PostedFile struct {
	Key      string
	FileName string
	Contents []byte
}
