package request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
)

const (
	// HTTPRequestLogLevelErrors writes only errors to the log.
	HTTPRequestLogLevelErrors = 1
	// HTTPRequestLogLevelVerbose writes lots of messages to the log.
	HTTPRequestLogLevelVerbose = 2
	//HTTPRequestLogLevelDebug writes more information to the log.
	HTTPRequestLogLevelDebug = 3
	// HTTPRequestLogLevelOver9000 writes everything to the log.
	HTTPRequestLogLevelOver9000 = 9001
)

//--------------------------------------------------------------------------------
// HttpResponseMeta
//--------------------------------------------------------------------------------

// NewHTTPResponseMeta returns a new meta object for a response.
func NewHTTPResponseMeta(res *http.Response) *HTTPResponseMeta {
	meta := &HTTPResponseMeta{}

	if res == nil {
		return meta
	}

	meta.StatusCode = res.StatusCode
	meta.ContentLength = res.ContentLength

	contentTypeHeader := res.Header["Content-Type"]
	if contentTypeHeader != nil && len(contentTypeHeader) > 0 {
		meta.ContentType = strings.Join(contentTypeHeader, ";")
	}

	contentEncodingHeader := res.Header["Content-Encoding"]
	if contentEncodingHeader != nil && len(contentEncodingHeader) > 0 {
		meta.ContentEncoding = strings.Join(contentEncodingHeader, ";")
	}

	meta.Headers = res.Header
	return meta
}

// HTTPRequestMeta is a summary of the request meta useful for logging.
type HTTPRequestMeta struct {
	Verb    string
	URL     *url.URL
	Headers http.Header
	Body    []byte
}

// HTTPResponseMeta is just the meta information for an http response.
type HTTPResponseMeta struct {
	StatusCode      int
	ContentLength   int64
	ContentEncoding string
	ContentType     string
	Headers         http.Header
}

// CreateTransportHandler is a receiver for `OnCreateTransport`.
type CreateTransportHandler func(host *url.URL, transport *http.Transport)

// ResponseHandler is a receiver for `OnResponse`.
type ResponseHandler func(meta *HTTPResponseMeta, content []byte)

// OutgoingRequestHandler is a receiver for `OnRequest`.
type OutgoingRequestHandler func(req *HTTPRequestMeta)

// MockedResponseHandler is a receiver for `WithMockedResponse`.
type MockedResponseHandler func(verb string, url *url.URL) (bool, *HTTPResponseMeta, []byte, error)

// Deserializer is a function that does things with the response body.
type Deserializer func(body []byte) error

// Serializer is a function that turns an object into raw data.
type Serializer func(value interface{}) ([]byte, error)

//--------------------------------------------------------------------------------
// HTTPRequest
//--------------------------------------------------------------------------------

// NewHTTPRequest returns a new HTTPRequest instance.
func NewHTTPRequest() *HTTPRequest {
	hr := HTTPRequest{}
	hr.Scheme = "http"
	hr.Verb = "GET"
	hr.KeepAlive = false
	return &hr
}

// HTTPRequest makes http requests.
type HTTPRequest struct {
	Scheme            string
	Host              string
	Path              string
	QueryString       url.Values
	Header            http.Header
	PostData          url.Values
	Cookies           []*http.Cookie
	BasicAuthUsername string
	BasicAuthPassword string
	Verb              string
	ContentType       string
	Timeout           time.Duration
	TLSCertPath       string
	TLSKeyPath        string
	Body              []byte
	KeepAlive         bool

	Label string

	Logger   *log.Logger
	LogLevel int

	transport *http.Transport

	createTransportHandler  CreateTransportHandler
	incomingResponseHandler ResponseHandler
	outgoingRequestHandler  OutgoingRequestHandler
	mockHandler             MockedResponseHandler
}

// OnResponse configures an event receiver.
func (hr *HTTPRequest) OnResponse(hook ResponseHandler) *HTTPRequest {
	hr.incomingResponseHandler = hook
	return hr
}

// OnCreateTransport configures an event receiver.
func (hr *HTTPRequest) OnCreateTransport(hook CreateTransportHandler) *HTTPRequest {
	hr.createTransportHandler = hook
	return hr
}

// OnRequest configures an event receiver.
func (hr *HTTPRequest) OnRequest(hook OutgoingRequestHandler) *HTTPRequest {
	hr.outgoingRequestHandler = hook
	return hr
}

// WithLabel gives the request a logging label.
func (hr *HTTPRequest) WithLabel(label string) *HTTPRequest {
	hr.Label = label
	return hr
}

// WithMockedResponse mocks a request response.
func (hr *HTTPRequest) WithMockedResponse(hook MockedResponseHandler) *HTTPRequest {
	hr.mockHandler = hook
	return hr
}

// WithLogging enables logging with HTTPRequestLogLevelErrors.
func (hr *HTTPRequest) WithLogging() *HTTPRequest {
	hr.LogLevel = HTTPRequestLogLevelErrors
	hr.Logger = log.New(os.Stdout, "", 0) // no error prefix
	return hr
}

// WithLogLevel sets a log level filter for the request.
func (hr *HTTPRequest) WithLogLevel(logLevel int) *HTTPRequest {
	hr.LogLevel = logLevel
	return hr
}

// WithLogger provides a logLevel and a logger for the request.
func (hr *HTTPRequest) WithLogger(logLevel int, logger *log.Logger) *HTTPRequest {
	hr.LogLevel = logLevel
	hr.Logger = logger
	return hr
}

func (hr *HTTPRequest) fatalf(logLevel int, format string, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		hr.Logger.Fatalf(prefix+format, args...)
	}
}

func (hr *HTTPRequest) fatal(logLevel int, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		message := fmt.Sprint(args...)
		fullMessage := fmt.Sprintf("%s%s", prefix, message)
		hr.Logger.Fatalln(fullMessage)
	}
}

func (hr *HTTPRequest) logf(logLevel int, format string, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		hr.Logger.Printf(prefix+format, args...)
	}
}

func (hr *HTTPRequest) log(logLevel int, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		message := fmt.Sprint(args...)
		fullMessage := fmt.Sprintf("%s%s", prefix, message)
		hr.Logger.Println(fullMessage)
	}
}

// WithTransport sets a transport for the request.
func (hr *HTTPRequest) WithTransport(transport *http.Transport) *HTTPRequest {
	hr.transport = transport
	return hr
}

// WithKeepAlives sets if the request should use the `Connection=keep-alive` header or not.
func (hr *HTTPRequest) WithKeepAlives() *HTTPRequest {
	hr.KeepAlive = true
	hr = hr.WithHeader("Connection", "keep-alive")
	return hr
}

// WithContentType sets the `Content-Type` header for the request.
func (hr *HTTPRequest) WithContentType(contentType string) *HTTPRequest {
	hr.ContentType = contentType
	return hr
}

// WithScheme sets the scheme, or protocol, of the request.
func (hr *HTTPRequest) WithScheme(scheme string) *HTTPRequest {
	hr.Scheme = scheme
	return hr
}

// WithHost sets the target url host for the request.
func (hr *HTTPRequest) WithHost(host string) *HTTPRequest {
	hr.Host = host
	return hr
}

// WithPath sets the path component of the host url..
func (hr *HTTPRequest) WithPath(path string) *HTTPRequest {
	hr.Path = path
	return hr
}

// WithPathf sets the path component of the host url by the format and arguments.
func (hr *HTTPRequest) WithPathf(format string, args ...interface{}) *HTTPRequest {
	hr.Path = fmt.Sprintf(format, args...)
	return hr
}

// WithCombinedPath sets the path component of the host url by combining the input path segments.
func (hr *HTTPRequest) WithCombinedPath(components ...string) *HTTPRequest {
	hr.Path = util.CombinePathComponents(components...)
	return hr
}

// WithURL sets the request target url whole hog.
func (hr *HTTPRequest) WithURL(urlString string) *HTTPRequest {
	workingURL, _ := url.Parse(urlString)
	hr.Scheme = workingURL.Scheme
	hr.Host = workingURL.Host
	hr.Path = workingURL.Path
	params := strings.Split(workingURL.RawQuery, "&")
	hr.QueryString = url.Values{}
	var keyValue []string
	for _, param := range params {
		if param != "" {
			keyValue = strings.Split(param, "=")
			hr.QueryString.Set(keyValue[0], keyValue[1])
		}
	}
	return hr
}

// WithHeader sets a header on the request.
func (hr *HTTPRequest) WithHeader(field string, value string) *HTTPRequest {
	if hr.Header == nil {
		hr.Header = http.Header{}
	}
	hr.Header.Set(field, value)
	return hr
}

// WithQueryString sets a query string value for the host url of the request.
func (hr *HTTPRequest) WithQueryString(field string, value string) *HTTPRequest {
	if hr.QueryString == nil {
		hr.QueryString = url.Values{}
	}
	hr.QueryString.Add(field, value)
	return hr
}

// WithCookie sets a cookie for the request.
func (hr *HTTPRequest) WithCookie(cookie *http.Cookie) *HTTPRequest {
	if hr.Cookies == nil {
		hr.Cookies = []*http.Cookie{}
	}
	hr.Cookies = append(hr.Cookies, cookie)
	return hr
}

// WithPostData sets a post data value for the request.
func (hr *HTTPRequest) WithPostData(field string, value string) *HTTPRequest {
	if hr.PostData == nil {
		hr.PostData = url.Values{}
	}
	hr.PostData.Add(field, value)
	return hr
}

// WithPostDataFromObject sets the post data for a request as json from a given object.
// Remarks; this differs from `WithJSONBody` in that it sets individual post form fields
// for each member of the object.
func (hr *HTTPRequest) WithPostDataFromObject(object interface{}) *HTTPRequest {
	postDatums := util.DecomposeToPostDataAsJSON(object)

	for _, item := range postDatums {
		hr.WithPostData(item.Key, item.Value)
	}

	return hr
}

// WithBasicAuth sets the basic auth headers for a request.
func (hr *HTTPRequest) WithBasicAuth(username, password string) *HTTPRequest {
	hr.BasicAuthUsername = username
	hr.BasicAuthPassword = password
	return hr
}

// WithTimeout sets a timeout for the request.
// Remarks: This timeout is enforced on client connect, not on request read + response.
func (hr *HTTPRequest) WithTimeout(timeout time.Duration) *HTTPRequest {
	hr.Timeout = timeout
	return hr
}

// WithTLSCert sets a tls cert on the transport for the request.
func (hr *HTTPRequest) WithTLSCert(certPath string) *HTTPRequest {
	hr.TLSCertPath = certPath
	return hr
}

// WithTLSKey sets a tls key on the transport for the request.
func (hr *HTTPRequest) WithTLSKey(keyPath string) *HTTPRequest {
	hr.TLSKeyPath = keyPath
	return hr
}

// WithVerb sets the http verb of the request.
func (hr *HTTPRequest) WithVerb(verb string) *HTTPRequest {
	hr.Verb = verb
	return hr
}

// AsGet sets the http verb of the request to `GET`.
func (hr *HTTPRequest) AsGet() *HTTPRequest {
	hr.Verb = "GET"
	return hr
}

// AsPost sets the http verb of the request to `POST`.
func (hr *HTTPRequest) AsPost() *HTTPRequest {
	hr.Verb = "POST"
	return hr
}

// AsPut sets the http verb of the request to `PUT`.
func (hr *HTTPRequest) AsPut() *HTTPRequest {
	hr.Verb = "PUT"
	return hr
}

// AsPatch sets the http verb of the request to `PATCH`.
func (hr *HTTPRequest) AsPatch() *HTTPRequest {
	hr.Verb = "PATCH"
	return hr
}

// AsDelete sets the http verb of the request to `DELETE`.
func (hr *HTTPRequest) AsDelete() *HTTPRequest {
	hr.Verb = "DELETE"
	return hr
}

// WithJSONBody sets the post body raw to be the json representation of an object.
func (hr *HTTPRequest) WithJSONBody(object interface{}) *HTTPRequest {
	return hr.WithSerializedBody(object, serializeJSON).WithContentType("application/json")
}

// WithXMLBody sets the post body raw to be the xml representation of an object.
func (hr *HTTPRequest) WithXMLBody(object interface{}) *HTTPRequest {
	return hr.WithSerializedBody(object, serializeXML).WithContentType("application/xml")
}

// WithBody sets the post body with the results of the given serializer.
func (hr *HTTPRequest) WithSerializedBody(object interface{}, serialize Serializer) *HTTPRequest {
	body, _ := serialize(object)
	return hr.WithRawBody(body)
}

// WithRawBody sets the post body directly.
func (hr *HTTPRequest) WithRawBody(body []byte) *HTTPRequest {
	hr.Body = body
	return hr
}

// CreateURL returns the currently formatted request target url.
func (hr *HTTPRequest) CreateURL() *url.URL {
	workingURL := &url.URL{Scheme: hr.Scheme, Host: hr.Host, Path: hr.Path}
	workingURL.RawQuery = hr.QueryString.Encode()
	return workingURL
}

func (hr *HTTPRequest) RequestMeta() *HTTPRequestMeta {
	return &HTTPRequestMeta{
		Verb:    hr.Verb,
		URL:     hr.CreateURL(),
		Body:    hr.RequestBody(),
		Headers: hr.Headers(),
	}
}

// RequestBody returns the current post body.
func (hr *HTTPRequest) RequestBody() []byte {
	if len(hr.Body) > 0 {
		return hr.Body
	} else if len(hr.PostData) > 0 {
		return []byte(hr.PostData.Encode())
	}
	return nil
}

func (hr *HTTPRequest) Headers() http.Header {
	headers := http.Header{}
	for key, values := range hr.Header {
		for _, value := range values {
			headers.Set(key, value)
		}
	}
	if len(hr.PostData) > 0 {
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if !isEmpty(hr.ContentType) {
		headers.Set("Content-Type", hr.ContentType)
	}
	return headers
}

// CreateHTTPRequest returns a http.Request for the HTTPRequest.
func (hr *HTTPRequest) CreateHTTPRequest() (*http.Request, error) {
	workingURL := hr.CreateURL()

	if len(hr.Body) > 0 && len(hr.PostData) > 0 {
		return nil, exception.New("Cant set both a body and have post data.")
	}

	req, err := http.NewRequest(hr.Verb, workingURL.String(), bytes.NewBuffer(hr.RequestBody()))
	if err != nil {
		return nil, exception.Wrap(err)
	}

	if !isEmpty(hr.BasicAuthUsername) {
		req.SetBasicAuth(hr.BasicAuthUsername, hr.BasicAuthPassword)
	}

	if hr.Cookies != nil {
		for i := 0; i < len(hr.Cookies); i++ {
			cookie := hr.Cookies[i]
			req.AddCookie(cookie)
		}
	}

	for key, values := range hr.Headers() {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	return req, nil
}

// FetchRawResponse makes the actual request but returns the underlying http.Response object.
func (hr *HTTPRequest) FetchRawResponse() (*http.Response, error) {
	req, reqErr := hr.CreateHTTPRequest()
	if reqErr != nil {
		return nil, reqErr
	}

	hr.logRequest()

	if hr.mockHandler != nil {
		didMockResponse, mockedMeta, mockedResponse, mockedResponseErr := hr.mockHandler(hr.Verb, req.URL)
		if didMockResponse {
			buff := bytes.NewBuffer(mockedResponse)
			res := http.Response{}
			buffLen := buff.Len()
			res.Body = ioutil.NopCloser(buff)
			res.ContentLength = int64(buffLen)
			res.Header = mockedMeta.Headers
			res.StatusCode = mockedMeta.StatusCode
			return &res, exception.Wrap(mockedResponseErr)
		}
	}

	client := &http.Client{}
	if hr.requiresCustomTransport() {
		transport, transportErr := hr.getHTTPTransport()
		if transportErr != nil {
			return nil, exception.Wrap(transportErr)
		}
		client.Transport = transport
	}

	if hr.Timeout != time.Duration(0) {
		client.Timeout = hr.Timeout
	}

	res, resErr := client.Do(req)
	return res, exception.Wrap(resErr)
}

// Execute makes the request but does not read the response.
func (hr *HTTPRequest) Execute() error {
	_, err := hr.ExecuteWithMeta()
	return exception.Wrap(err)
}

// ExecuteWithMeta makes the request and returns the meta of the response.
func (hr *HTTPRequest) ExecuteWithMeta() (*HTTPResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	if res != nil && res.Body != nil {
		closeErr := res.Body.Close()
		if closeErr != nil {
			return nil, exception.WrapMany(exception.Wrap(err), exception.Wrap(closeErr))
		}
	}
	meta := NewHTTPResponseMeta(res)
	return meta, exception.Wrap(err)
}

// FetchString returns the body of the response as a string.
func (hr *HTTPRequest) FetchString() (string, error) {
	responseStr, _, err := hr.FetchStringWithMeta()
	return responseStr, err
}

// FetchStringWithMeta returns the body of the response as a string in addition to the response metadata.
func (hr *HTTPRequest) FetchStringWithMeta() (string, *HTTPResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := NewHTTPResponseMeta(res)
	if err != nil {
		return util.StringEmpty, meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	bytes, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return util.StringEmpty, meta, exception.Wrap(readErr)
	}

	meta.ContentLength = int64(len(bytes))
	hr.logResponse(meta, bytes)
	return string(bytes), meta, nil
}

// FetchJSONToObject unmarshals the response as json to an object.
func (hr *HTTPRequest) FetchJSONToObject(destination interface{}) error {
	_, err := hr.deserialize(newJSONDeserializer(destination))
	return err
}

// FetchJSONToObjectWithMeta unmarshals the response as json to an object with metadata.
func (hr *HTTPRequest) FetchJSONToObjectWithMeta(destination interface{}) (*HTTPResponseMeta, error) {
	return hr.deserialize(newJSONDeserializer(destination))
}

// FetchJSONToObjectWithErrorHandler unmarshals the response as json to an object with metadata or an error object depending on the meta.
func (hr *HTTPRequest) FetchJSONToObjectWithErrorHandler(successObject interface{}, errorObject interface{}) (*HTTPResponseMeta, error) {
	return hr.deserializeWithError(newJSONDeserializer(successObject), newJSONDeserializer(errorObject))
}

// FetchJSONError unmarshals the response as json to an object if the meta indiciates an error.
func (hr *HTTPRequest) FetchJSONError(errorObject interface{}) (*HTTPResponseMeta, error) {
	return hr.deserializeWithError(nil, newJSONDeserializer(errorObject))
}

// FetchXMLToObject unmarshals the response as xml to an object with metadata.
func (hr *HTTPRequest) FetchXMLToObject(destination interface{}) error {
	_, err := hr.deserialize(newXMLDeserializer(destination))
	return err
}

// FetchXMLToObjectWithMeta unmarshals the response as xml to an object with metadata.
func (hr *HTTPRequest) FetchXMLToObjectWithMeta(destination interface{}) (*HTTPResponseMeta, error) {
	return hr.deserialize(newXMLDeserializer(destination))
}

// FetchXMLToObjectWithErrorHandler unmarshals the response as xml to an object with metadata or an error object depending on the meta.
func (hr *HTTPRequest) FetchXMLToObjectWithErrorHandler(successObject interface{}, errorObject interface{}) (*HTTPResponseMeta, error) {
	return hr.deserializeWithError(newXMLDeserializer(successObject), newXMLDeserializer(errorObject))
}

// FetchObjectWithSerializer runs a deserializer with the response.
func (hr *HTTPRequest) FetchObjectWithSerializer(deserialize Deserializer) (*HTTPResponseMeta, error) {
	meta, responseErr := hr.deserialize(func(body []byte) error {
		return deserialize(body)
	})
	return meta, responseErr
}

func (hr *HTTPRequest) requiresCustomTransport() bool {
	return (!isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath)) || hr.transport != nil || hr.createTransportHandler != nil
}

func (hr *HTTPRequest) getHTTPTransport() (*http.Transport, error) {
	if hr.transport != nil {
		hr.log(HTTPRequestLogLevelDebug, "Service Request ==> Using Provided Transport\n")
		return hr.transport, nil
	}
	return hr.createHTTPTransport()
}

func (hr *HTTPRequest) createHTTPTransport() (*http.Transport, error) {
	hr.log(HTTPRequestLogLevelDebug, "Service Request ==> Creating Custom Transport\n")
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  !hr.KeepAlive,
	}

	dialer := &net.Dialer{}
	if hr.Timeout != time.Duration(0) {
		dialer.Timeout = hr.Timeout
	}
	if hr.KeepAlive {
		hr.logf(HTTPRequestLogLevelDebug, "Service Request ==> Transport Enabled For `keep-alive` %v\n", 30*time.Second)
		dialer.KeepAlive = 30 * time.Second
	}

	loggedDialer := func(network, address string) (net.Conn, error) {
		hr.logf(HTTPRequestLogLevelDebug, "Service Request ==> Transport Is Dialing %s\n", address)
		return dialer.Dial(network, address)
	}
	transport.Dial = loggedDialer

	if !isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath) {
		cert, err := tls.LoadX509KeyPair(hr.TLSCertPath, hr.TLSKeyPath)
		if err != nil {
			return nil, exception.Wrap(err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		transport.TLSClientConfig = tlsConfig
	}

	if hr.createTransportHandler != nil {
		hr.createTransportHandler(hr.CreateURL(), transport)
	}

	return transport, nil
}

func (hr *HTTPRequest) deserialize(handler Deserializer) (*HTTPResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := NewHTTPResponseMeta(res)

	if err != nil {
		return meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return meta, exception.Wrap(err)
	}

	meta.ContentLength = int64(len(body))
	hr.logResponse(meta, body)
	if handler != nil {
		err = handler(body)
	}
	return meta, exception.Wrap(err)
}

func (hr *HTTPRequest) deserializeWithError(okHandler Deserializer, errorHandler Deserializer) (*HTTPResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := NewHTTPResponseMeta(res)

	if err != nil {
		return meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return meta, exception.Wrap(err)
	}

	meta.ContentLength = int64(len(body))
	hr.logResponse(meta, body)
	if res.StatusCode == http.StatusOK {
		if okHandler != nil {
			err = okHandler(body)
		}
	} else if errorHandler != nil {
		err = errorHandler(body)
	}
	return meta, exception.Wrap(err)
}

func (hr *HTTPRequest) logRequest() {
	meta := hr.RequestMeta()
	if hr.outgoingRequestHandler != nil {
		hr.outgoingRequestHandler(meta)
	}
	hr.logf(HTTPRequestLogLevelVerbose, "Service Request ==> %s %s\n", meta.Verb, meta.URL.String())
}

func (hr *HTTPRequest) logResponse(meta *HTTPResponseMeta, responseBody []byte) {
	if hr.incomingResponseHandler != nil {
		hr.incomingResponseHandler(meta, responseBody)
	}
	hr.logf(HTTPRequestLogLevelVerbose, "Service Response ==> %s", responseBody)
}

//--------------------------------------------------------------------------------
// Unexported Utility Functions
//--------------------------------------------------------------------------------

func newJSONDeserializer(object interface{}) Deserializer {
	return func(body []byte) error {
		return deserializeJSON(object, body)
	}
}

func newXMLDeserializer(object interface{}) Deserializer {
	return func(body []byte) error {
		return deserializeXML(object, body)
	}
}

func deserializeJSON(object interface{}, body []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(body))
	decodeErr := decoder.Decode(object)
	return exception.Wrap(decodeErr)
}

func deserializeJSONFromReader(object interface{}, body io.Reader) error {
	decoder := json.NewDecoder(body)
	decodeErr := decoder.Decode(object)
	return exception.Wrap(decodeErr)
}

func serializeJSON(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

func serializeJSONToReader(object interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(object)
	return buf, err
}

func deserializeXML(object interface{}, body []byte) error {
	return deserializeXMLFromReader(object, bytes.NewBuffer(body))
}

func deserializeXMLFromReader(object interface{}, reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	return decoder.Decode(object)
}

func serializeXML(object interface{}) ([]byte, error) {
	return xml.Marshal(object)
}

func serializeXMLToReader(object interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := xml.NewEncoder(buf)
	err := encoder.Encode(object)
	return buf, err
}

func getLoggingPrefix(logLevel int) string {
	return fmt.Sprintf("HttpRequest (%s): ", formatLogLevel(logLevel))
}

func formatLogLevel(logLevel int) string {
	switch logLevel {
	case HTTPRequestLogLevelErrors:
		return "ERRORS"
	case HTTPRequestLogLevelVerbose:
		return "VERBOSE"
	case HTTPRequestLogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

func isEmpty(str string) bool {
	return len(str) == 0
}
