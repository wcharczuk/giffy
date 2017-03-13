package request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util"
)

//--------------------------------------------------------------------------------
// HTTPRequestMeta
//--------------------------------------------------------------------------------

// NewHTTPRequestMeta returns a new meta object for a request.
func NewHTTPRequestMeta(req *http.Request) *HTTPRequestMeta {
	return &HTTPRequestMeta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
	}
}

// NewHTTPRequestMetaWithBody returns a new meta object for a request and reads the body.
func NewHTTPRequestMetaWithBody(req *http.Request) (*HTTPRequestMeta, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return &HTTPRequestMeta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
		Body:    body,
	}, nil
}

// HTTPRequestMeta is a summary of the request meta useful for logging.
type HTTPRequestMeta struct {
	StartTime time.Time
	Verb      string
	URL       *url.URL
	Headers   http.Header
	Body      []byte
}

//--------------------------------------------------------------------------------
// HttpResponseMeta
//--------------------------------------------------------------------------------

// NewHTTPResponseMeta returns a new meta object for a response.
func NewHTTPResponseMeta(res *http.Response) *HTTPResponseMeta {
	meta := &HTTPResponseMeta{}

	if res == nil {
		return meta
	}

	meta.CompleteTime = time.Now().UTC()
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

// HTTPResponseMeta is just the meta information for an http response.
type HTTPResponseMeta struct {
	CompleteTime    time.Time
	StatusCode      int
	ContentLength   int64
	ContentEncoding string
	ContentType     string
	Headers         http.Header
}

// CreateTransportHandler is a receiver for `OnCreateTransport`.
type CreateTransportHandler func(host *url.URL, transport *http.Transport)

// ResponseHandler is a receiver for `OnResponse`.
type ResponseHandler func(req *HTTPRequestMeta, meta *HTTPResponseMeta, content []byte)

// StatefulResponseHandler is a receiver for `OnResponse` that includes a state object.
type StatefulResponseHandler func(req *HTTPRequestMeta, res *HTTPResponseMeta, content []byte, state interface{})

// OutgoingRequestHandler is a receiver for `OnRequest`.
type OutgoingRequestHandler func(req *HTTPRequestMeta)

// MockedResponseHandler is a receiver for `WithMockedResponse`.
type MockedResponseHandler func(verb string, url *url.URL) (bool, *HTTPResponseMeta, []byte, error)

// Deserializer is a function that does things with the response body.
type Deserializer func(body []byte) error

// Serializer is a function that turns an object into raw data.
type Serializer func(value interface{}) ([]byte, error)

//--------------------------------------------------------------------------------
// PostedFile
//--------------------------------------------------------------------------------

// PostedFile represents a file to post with the request.
type PostedFile struct {
	Key          string
	FileName     string
	FileContents io.Reader
}

//--------------------------------------------------------------------------------
// Buffer
//--------------------------------------------------------------------------------

// Buffer is a type that supplies two methods found on bytes.Buffer.
type Buffer interface {
	Write([]byte) (int, error)
	Len() int64
	ReadFrom(io.ReadCloser) (int64, error)
	Bytes() []byte
}

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
	Scheme              string
	Host                string
	Path                string
	QueryString         url.Values
	Header              http.Header
	PostData            url.Values
	Cookies             []*http.Cookie
	BasicAuthUsername   string
	BasicAuthPassword   string
	Verb                string
	ContentType         string
	Timeout             time.Duration
	TLSCertPath         string
	TLSKeyPath          string
	SkipTLSVerification bool
	Body                []byte
	KeepAlive           bool

	Label string

	logger *logger.Agent

	state interface{}

	postedFiles []PostedFile

	responseBuffer Buffer

	requestStart time.Time

	transport                       *http.Transport
	createTransportHandler          CreateTransportHandler
	incomingResponseHandler         ResponseHandler
	statefulIncomingResponseHandler StatefulResponseHandler
	outgoingRequestHandler          OutgoingRequestHandler
	mockHandler                     MockedResponseHandler
}

// OnResponse configures an event receiver.
func (hr *HTTPRequest) OnResponse(hook ResponseHandler) *HTTPRequest {
	hr.incomingResponseHandler = hook
	return hr
}

// OnResponseStateful configures an event receiver that includes the request state.
func (hr *HTTPRequest) OnResponseStateful(hook StatefulResponseHandler) *HTTPRequest {
	hr.statefulIncomingResponseHandler = hook
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

// WithState adds a state object to the request for later usage.
func (hr *HTTPRequest) WithState(state interface{}) *HTTPRequest {
	hr.state = state
	return hr
}

// WithLabel gives the request a logging label.
func (hr *HTTPRequest) WithLabel(label string) *HTTPRequest {
	hr.Label = label
	return hr
}

// ShouldSkipTLSVerification skips the bad certificate checking on TLS requests.
func (hr *HTTPRequest) ShouldSkipTLSVerification() *HTTPRequest {
	hr.SkipTLSVerification = true
	return hr
}

// WithMockedResponse mocks a request response.
func (hr *HTTPRequest) WithMockedResponse(hook MockedResponseHandler) *HTTPRequest {
	hr.mockHandler = hook
	return hr
}

// WithLogger enables logging with HTTPRequestLogLevelErrors.
func (hr *HTTPRequest) WithLogger(agent *logger.Agent) *HTTPRequest {
	hr.logger = agent
	return hr
}

// Logger returns the request diagnostics agent.
func (hr *HTTPRequest) Logger() *logger.Agent {
	return hr.logger
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
	hr.Path = util.String.CombinePathComponents(components...)
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
	postDatums := util.Reflection.DecomposeToPostDataAsJSON(object)

	for _, item := range postDatums {
		hr.WithPostData(item.Key, item.Value)
	}

	return hr
}

// WithPostedFile adds a posted file to the multipart form elements of the request.
func (hr *HTTPRequest) WithPostedFile(key, fileName string, fileContents io.Reader) *HTTPRequest {
	hr.postedFiles = append(hr.postedFiles, PostedFile{Key: key, FileName: fileName, FileContents: fileContents})
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

// WithResponseBuffer sets the response buffer for the request (if you want to re-use one).
func (hr *HTTPRequest) WithResponseBuffer(buffer Buffer) *HTTPRequest {
	hr.responseBuffer = buffer
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

// WithSerializedBody sets the post body with the results of the given serializer.
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

// AsRequestMeta returns the request as a HTTPRequestMeta.
func (hr *HTTPRequest) AsRequestMeta() *HTTPRequestMeta {
	return &HTTPRequestMeta{
		StartTime: hr.requestStart,
		Verb:      hr.Verb,
		URL:       hr.CreateURL(),
		Body:      hr.RequestBody(),
		Headers:   hr.Headers(),
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

// Headers returns the headers on the request.
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
	if err != nil {
		return nil, exception.Wrap(err)
	}
	meta := NewHTTPResponseMeta(res)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
		if hr.responseBuffer != nil {
			contentLength, err := hr.responseBuffer.ReadFrom(res.Body)
			if err != nil {
				return nil, exception.Wrap(err)
			}
			meta.ContentLength = contentLength
			if hr.incomingResponseHandler != nil {
				hr.logResponse(meta, hr.responseBuffer.Bytes(), hr.state)
			}
		} else {
			contents, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, exception.Wrap(err)
			}
			meta.ContentLength = int64(len(contents))
			hr.logResponse(meta, contents, hr.state)
		}
	}

	return meta, nil
}

// FetchBytesWithMeta fetches the response as bytes with meta.
func (hr *HTTPRequest) FetchBytesWithMeta() ([]byte, *HTTPResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	resMeta := NewHTTPResponseMeta(res)
	if err != nil {
		return nil, resMeta, exception.Wrap(err)
	}
	defer res.Body.Close()

	bytes, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, resMeta, exception.Wrap(readErr)
	}

	resMeta.ContentLength = int64(len(bytes))
	hr.logResponse(resMeta, bytes, hr.state)
	return bytes, resMeta, nil
}

// FetchBytes fetches the response as bytes.
func (hr *HTTPRequest) FetchBytes() ([]byte, error) {
	contents, _, err := hr.FetchBytesWithMeta()
	return contents, err
}

// FetchString returns the body of the response as a string.
func (hr *HTTPRequest) FetchString() (string, error) {
	responseStr, _, err := hr.FetchStringWithMeta()
	return responseStr, err
}

// FetchStringWithMeta returns the body of the response as a string in addition to the response metadata.
func (hr *HTTPRequest) FetchStringWithMeta() (string, *HTTPResponseMeta, error) {
	contents, meta, err := hr.FetchBytesWithMeta()
	return string(contents), meta, err
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
	return (!isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath)) ||
		hr.transport != nil ||
		hr.createTransportHandler != nil ||
		hr.SkipTLSVerification
}

func (hr *HTTPRequest) getHTTPTransport() (*http.Transport, error) {
	if hr.transport != nil {
		return hr.transport, nil
	}
	return hr.CreateHTTPTransport()
}

// CreateHTTPTransport returns the the custom transport for the request.
func (hr *HTTPRequest) CreateHTTPTransport() (*http.Transport, error) {
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  !hr.KeepAlive,
	}

	dialer := &net.Dialer{}
	if hr.Timeout != time.Duration(0) {
		dialer.Timeout = hr.Timeout
	}
	if hr.KeepAlive {
		dialer.KeepAlive = 30 * time.Second
	}

	loggedDialer := func(network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}
	transport.Dial = loggedDialer

	if !isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath) {
		cert, err := tls.LoadX509KeyPair(hr.TLSCertPath, hr.TLSKeyPath)
		if err != nil {
			return nil, exception.Wrap(err)
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: hr.SkipTLSVerification,
			Certificates:       []tls.Certificate{cert},
		}
		transport.TLSClientConfig = tlsConfig
	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: hr.SkipTLSVerification,
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
	hr.logResponse(meta, body, hr.state)
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
	hr.logResponse(meta, body, hr.state)
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
	hr.requestStart = time.Now().UTC()

	meta := hr.AsRequestMeta()
	if hr.outgoingRequestHandler != nil {
		hr.outgoingRequestHandler(meta)
	}

	if hr.logger != nil {
		hr.logger.OnEvent(Event, meta)
	}
}

func (hr *HTTPRequest) logResponse(resMeta *HTTPResponseMeta, responseBody []byte, state interface{}) {
	if hr.statefulIncomingResponseHandler != nil {
		hr.statefulIncomingResponseHandler(hr.AsRequestMeta(), resMeta, responseBody, state)
	}
	if hr.incomingResponseHandler != nil {
		hr.incomingResponseHandler(hr.AsRequestMeta(), resMeta, responseBody)
	}

	if hr.logger != nil {
		hr.logger.OnEvent(EventResponse, hr.AsRequestMeta(), resMeta, responseBody, state)
	}
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

func isEmpty(str string) bool {
	return len(str) == 0
}
