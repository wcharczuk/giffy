package web

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"strings"

	logger "github.com/blendlabs/go-logger"
)

const (
	// PostBodySize is the maximum post body size we will typically consume.
	PostBodySize = int64(1 << 26) //64mb

	// PostBodySizeMax is the absolute maximum file size the server can handle.
	PostBodySizeMax = int64(1 << 32) //enormous.

	// StringEmpty is the empty string.
	StringEmpty = ""
)

// NewCtx returns a new hc context.
func NewCtx(w ResponseWriter, r *http.Request, p RouteParameters) *Ctx {
	ctx := &Ctx{
		Response:        w,
		Request:         r,
		routeParameters: p,
		state:           State{},
	}

	return ctx
}

// Ctx is the struct that represents the context for an hc request.
type Ctx struct {
	//Public fields
	Response ResponseWriter
	Request  *http.Request

	app    *App
	logger *logger.Agent
	auth   *AuthManager

	postBody []byte

	//Private fields
	view                  *ViewResultProvider
	api                   *APIResultProvider
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

	tx *sql.Tx
}

// WithTx sets a transaction on the context.
func (rc *Ctx) WithTx(tx *sql.Tx) *Ctx {
	rc.tx = tx
	return rc
}

// Tx returns the transaction for the request.
func (rc *Ctx) Tx() *sql.Tx {
	return rc.tx
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

// Auth returns the AuthManager for the request.
func (rc *Ctx) Auth() *AuthManager {
	return rc.auth
}

// SetAuth sets the request context auth.
func (rc *Ctx) SetAuth(authManager *AuthManager) {
	rc.auth = authManager
}

// Session returns the session (if any) on the request.
func (rc *Ctx) Session() *Session {
	if rc.session != nil {
		return rc.session
	}

	return nil
}

// SetSession sets the session for the request.
func (rc *Ctx) SetSession(session *Session) {
	rc.session = session
}

// View returns the view result provider.
func (rc *Ctx) View() *ViewResultProvider {
	if rc.view == nil {
		rc.view = NewViewResultProvider(rc, rc.app.viewCache)
	}
	return rc.view
}

// API returns the view result provider.
func (rc *Ctx) API() *APIResultProvider {
	if rc.api == nil {
		rc.api = NewAPIResultProvider(rc)
	}
	return rc.api
}

// JSON returns the JSON result provider.
func (rc *Ctx) JSON() *JSONResultProvider {
	if rc.json == nil {
		rc.json = NewJSONResultProvider(rc)
	}
	return rc.json
}

// XML returns the xml result provider.
func (rc *Ctx) XML() *XMLResultProvider {
	if rc.xml == nil {
		rc.xml = NewXMLResultProvider(rc)
	}
	return rc.xml
}

// Text returns the text result provider.
func (rc *Ctx) Text() *TextResultProvider {
	if rc.text == nil {
		rc.text = NewTextResultProvider(rc)
	}
	return rc.text
}

// DefaultResultProvider returns the current result provider for the context. This is
// set by calling SetDefaultResultProvider or using one of the pre-built middleware
// steps that set it for you.
func (rc *Ctx) DefaultResultProvider() ResultProvider {
	if rc.defaultResultProvider == nil {
		rc.defaultResultProvider = NewTextResultProvider(rc)
	}
	return rc.defaultResultProvider
}

// SetDefaultResultProvider sets the current result provider.
func (rc *Ctx) SetDefaultResultProvider(provider ResultProvider) {
	rc.defaultResultProvider = provider
}

// State returns an object in the state cache.
func (rc *Ctx) State(key string) interface{} {
	if item, hasItem := rc.state[key]; hasItem {
		return item
	}
	return nil
}

// SetState sets the state for a key to an object.
func (rc *Ctx) SetState(key string, value interface{}) {
	rc.state[key] = value
}

// Param returns a parameter from the request.
func (rc *Ctx) Param(name string) string {
	if rc.routeParameters != nil {
		routeValue := rc.routeParameters.Get(name)
		if len(routeValue) > 0 {
			return routeValue
		}
	}
	if rc.Request != nil {
		if rc.Request.URL != nil {
			queryValue := rc.Request.URL.Query().Get(name)
			if len(queryValue) > 0 {
				return queryValue
			}
		}
		if rc.Request.Header != nil {
			headerValue := rc.Request.Header.Get(name)
			if len(headerValue) > 0 {
				return headerValue
			}
		}

		formValue := rc.Request.FormValue(name)
		if len(formValue) > 0 {
			return formValue
		}

		cookie, cookieErr := rc.Request.Cookie(name)
		if cookieErr == nil && len(cookie.Value) != 0 {
			return cookie.Value
		}
	}

	return ""
}

// ParamInt returns a parameter from any location as an integer.
func (rc *Ctx) ParamInt(name string) (int, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, newParameterMissingError(name)
	}
	return strconv.Atoi(paramValue)
}

// ParamInt64 returns a parameter from any location as an int64.
func (rc *Ctx) ParamInt64(name string) (int64, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, newParameterMissingError(name)
	}
	return strconv.ParseInt(paramValue, 10, 64)
}

// ParamFloat64 returns a parameter from any location as a float64.
func (rc *Ctx) ParamFloat64(name string) (float64, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, newParameterMissingError(name)
	}
	return strconv.ParseFloat(paramValue, 64)
}

// ParamTime returns a parameter from any location as a time with a given format.
func (rc *Ctx) ParamTime(name, format string) (time.Time, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return time.Time{}, newParameterMissingError(name)
	}
	return time.Parse(format, paramValue)
}

// ParamBool returns a boolean value for a param.
func (rc *Ctx) ParamBool(name string) (bool, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return false, newParameterMissingError(name)
	}
	lower := strings.ToLower(paramValue)
	return lower == "true" || lower == "1" || lower == "yes", nil
}

func (rc *Ctx) onPostBody(bodyContents []byte) {
	if rc.logger != nil {
		rc.logger.OnEvent(logger.EventWebRequestPostBody, rc.postBody)
	}
}

// PostBody returns the bytes in a post body.
func (rc *Ctx) PostBody() ([]byte, error) {
	var err error
	if len(rc.postBody) == 0 {
		defer rc.Request.Body.Close()
		rc.postBody, err = ioutil.ReadAll(rc.Request.Body)
		if err != nil {
			return nil, err
		}
		rc.onPostBody(rc.postBody)
	}

	return rc.postBody, err
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
			files = append(files, PostedFile{Key: key, FileName: fileHeader.Filename, Contents: bytes})
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
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, newParameterMissingError(key)
}

// QueryParamInt returns a query parameter as an integer.
func (rc *Ctx) QueryParamInt(key string) (int, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamInt64 returns a query parameter as an int64.
func (rc *Ctx) QueryParamInt64(key string) (int64, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamFloat64 returns a query parameter as a float64.
func (rc *Ctx) QueryParamFloat64(key string) (float64, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, newParameterMissingError(key)
}

// QueryParamTime returns a query parameter as a time.Time.
func (rc *Ctx) QueryParamTime(key, format string) (time.Time, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return time.Parse(format, value)
	}
	return time.Time{}, newParameterMissingError(key)
}

// HeaderParam returns a header parameter value.
func (rc *Ctx) HeaderParam(key string) (string, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, newParameterMissingError(key)
}

// HeaderParamInt returns a header parameter value as an integer.
func (rc *Ctx) HeaderParamInt(key string) (int, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamInt64 returns a header parameter value as an integer.
func (rc *Ctx) HeaderParamInt64(key string) (int64, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamFloat64 returns a header parameter value as an float64.
func (rc *Ctx) HeaderParamFloat64(key string) (float64, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, newParameterMissingError(key)
}

// HeaderParamTime returns a header parameter value as an float64.
func (rc *Ctx) HeaderParamTime(key, format string) (time.Time, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return time.Parse(format, key)
	}
	return time.Time{}, newParameterMissingError(key)
}

// GetCookie returns a named cookie from the request.
func (rc *Ctx) GetCookie(name string) *http.Cookie {
	cookie, err := rc.Request.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}

// WriteCookie writes the cookie to the response.
func (rc *Ctx) WriteCookie(cookie *http.Cookie) {
	http.SetCookie(rc.Response, cookie)
}

func (rc *Ctx) getCookieDomain() string {
	if rc.app != nil && rc.app.baseURL != nil {
		return rc.app.baseURL.Host
	}
	return rc.Request.Host
}

// WriteNewCookie is a helper method for WriteCookie.
func (rc *Ctx) WriteNewCookie(name string, value string, expires *time.Time, path string, secure bool) {
	c := http.Cookie{}
	c.Name = name
	c.HttpOnly = true
	c.Value = value
	c.Path = path
	c.Secure = secure
	c.Domain = rc.getCookieDomain()
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
func (rc *Ctx) Logger() *logger.Agent {
	return rc.logger
}

func (rc *Ctx) logFatal(err error) {
	if rc.logger != nil {
		rc.logger.ErrorEventWithState(logger.EventFatalError, logger.ColorRed, err, rc)
	}
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

// RawJSON returns a basic json result.
func (rc *Ctx) RawJSON(object interface{}) *JSONResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   object,
	}
}

// RawXML returns a basic xml result.
func (rc *Ctx) RawXML(object interface{}) *XMLResult {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   object,
	}
}

// NoContent returns a service response.
func (rc *Ctx) NoContent() *NoContentResult {
	return &NoContentResult{}
}

// Static returns a static result.
func (rc *Ctx) Static(filePath string) *StaticResult {
	return NewStaticResultForSingleFile(filePath)
}

// Redirect returns a redirect result.
func (rc *Ctx) Redirect(path string) *RedirectResult {
	return &RedirectResult{
		RedirectURI: path,
	}
}

// Redirectf returns a redirect result.
func (rc *Ctx) Redirectf(format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		RedirectURI: fmt.Sprintf(format, args...),
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

// Reset resets the context after handling a request.
func (rc *Ctx) Reset() {
	rc.app = nil
	rc.Request = nil
	rc.Response = nil
	rc.postBody = nil
	rc.route = nil
	rc.routeParameters = nil
	rc.session = nil
	rc.state = nil
	rc.statusCode = 0
	rc.contentLength = 0
	rc.requestStart = time.Time{}
	rc.requestEnd = time.Time{}
	rc.tx = nil
}

// PostedFile is a file that has been posted to an hc endpoint.
type PostedFile struct {
	Key      string
	FileName string
	Contents []byte
}

// State is the collection of state objects on a context.
type State map[string]interface{}
