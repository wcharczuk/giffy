package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/blendlabs/go-exception"
)

// NewMockRequestBuilder returns a new mock request builder for a given app.
func NewMockRequestBuilder(app *App) *MockRequestBuilder {
	return &MockRequestBuilder{
		app:         app,
		verb:        "GET",
		queryString: url.Values{},
		formValues:  url.Values{},
		headers:     http.Header{},
	}
}

// MockRequestBuilder facilitates creating mock requests.
type MockRequestBuilder struct {
	app *App

	verb        string
	path        string
	queryString url.Values
	formValues  url.Values
	headers     http.Header
	cookies     []*http.Cookie
	postBody    []byte

	responseBuffer *bytes.Buffer
}

// WithPathf sets the path for the request.
func (mrb *MockRequestBuilder) WithPathf(pathFormat string, args ...interface{}) *MockRequestBuilder {
	mrb.path = fmt.Sprintf(pathFormat, args...)
	return mrb
}

// WithVerb sets the verb for the request.
func (mrb *MockRequestBuilder) WithVerb(verb string) *MockRequestBuilder {
	mrb.verb = strings.ToUpper(verb)
	return mrb
}

// WithQueryString adds a querystring param for the request.
func (mrb *MockRequestBuilder) WithQueryString(key, value string) *MockRequestBuilder {
	mrb.queryString.Add(key, value)
	return mrb
}

// WithFormValue adds a form value for the request.
func (mrb *MockRequestBuilder) WithFormValue(key, value string) *MockRequestBuilder {
	mrb.formValues.Add(key, value)
	return mrb
}

// WithHeader adds a header for the request.
func (mrb *MockRequestBuilder) WithHeader(key, value string) *MockRequestBuilder {
	mrb.headers.Add(key, value)
	return mrb
}

// WithCookie adds a cookie for the request.
func (mrb *MockRequestBuilder) WithCookie(cookie *http.Cookie) *MockRequestBuilder {
	mrb.cookies = append(mrb.cookies, cookie)
	return mrb
}

// WithPostBody sets the post body for the request.
func (mrb *MockRequestBuilder) WithPostBody(postBody []byte) *MockRequestBuilder {
	mrb.postBody = postBody
	return mrb
}

// WithPostBodyAsJSON sets the post body for the request by serializing an object to JSON.
func (mrb *MockRequestBuilder) WithPostBodyAsJSON(object interface{}) *MockRequestBuilder {
	bytes, _ := json.Marshal(object)
	mrb.postBody = bytes
	return mrb
}

// WithResponseBuffer optionally sets a response buffer to write to during Execute and AsRequestContext.
func (mrb *MockRequestBuilder) WithResponseBuffer(buffer *bytes.Buffer) *MockRequestBuilder {
	mrb.responseBuffer = buffer
	return mrb
}

// AsRequest returns the mock request builder settings as an http.Request.
func (mrb *MockRequestBuilder) AsRequest() (*http.Request, error) {
	req := &http.Request{}
	reqURL, err := url.Parse(fmt.Sprintf("http://localhost/%s", mrb.path))
	if err != nil {
		return nil, err
	}

	reqURL.RawQuery = mrb.queryString.Encode()

	req.Method = mrb.verb
	req.URL = reqURL
	req.RequestURI = reqURL.String()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(mrb.postBody))
	req.Form = mrb.formValues
	req.Header = http.Header{}

	for key, values := range mrb.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for _, cookie := range mrb.cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

// AsRequestContext returns the mock request as a request context.
func (mrb *MockRequestBuilder) AsRequestContext(p RouteParameters) (*RequestContext, error) {
	r, err := mrb.AsRequest()

	if err != nil {
		return nil, err
	}

	var buffer *bytes.Buffer
	if mrb.responseBuffer != nil {
		buffer = mrb.responseBuffer
	} else {
		buffer = bytes.NewBuffer([]byte{})
	}

	w := NewMockResponseWriter(buffer)
	var rc *RequestContext
	if mrb.app != nil {
		rc = mrb.app.requestContext(w, r, p)
	} else {
		rc = NewRequestContext(w, r, p)
	}

	return rc, nil
}

// Response runs the mock request.
func (mrb *MockRequestBuilder) Response() (*http.Response, error) {
	handle, params, addTrailingSlash := mrb.app.router.Lookup(mrb.verb, mrb.path)
	if addTrailingSlash {
		mrb.path = mrb.path + "/"
	}

	handle, params, addTrailingSlash = mrb.app.router.Lookup(mrb.verb, mrb.path)
	if handle == nil {
		return nil, exception.Newf("No matching route for path %s `%s`", mrb.verb, mrb.path)
	}

	req, err := mrb.AsRequest()
	if err != nil {
		return nil, err
	}

	var buffer *bytes.Buffer
	if mrb.responseBuffer != nil {
		buffer = mrb.responseBuffer
	} else {
		buffer = bytes.NewBuffer([]byte{})
	}

	w := NewMockResponseWriter(buffer)
	handle(w, req, params)
	res := http.Response{
		Body:          ioutil.NopCloser(bytes.NewBuffer(buffer.Bytes())),
		ContentLength: int64(w.ContentLength()),
		Header:        http.Header{},
	}

	for key, values := range w.Header() {
		for _, value := range values {
			res.Header.Add(key, value)
		}
	}

	res.StatusCode = w.statusCode
	res.Proto = "http"
	res.ProtoMajor = 1
	res.ProtoMinor = 1

	return &res, nil
}

// JSON executes the mock request and reads the response to the given object as json.
func (mrb *MockRequestBuilder) JSON(object interface{}) error {
	res, err := mrb.Response()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, object)
}

// Bytes returns the response as bytes.
func (mrb *MockRequestBuilder) Bytes() ([]byte, error) {
	res, err := mrb.Response()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// Execute just runs the request.
func (mrb *MockRequestBuilder) Execute() error {
	_, err := mrb.Bytes()
	return err
}
