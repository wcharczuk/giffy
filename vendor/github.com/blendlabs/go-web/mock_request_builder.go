package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/blendlabs/go-exception"
)

// NewMockRequestBuilder returns a new mock request builder for a given app.
func NewMockRequestBuilder(app *App) *MockRequestBuilder {
	var err error
	if app != nil {
		err = app.commonStartupTasks()
	}

	return &MockRequestBuilder{
		app:         app,
		verb:        "GET",
		queryString: url.Values{},
		formValues:  url.Values{},
		headers:     http.Header{},
		startupErr:  err,
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

	postedFiles map[string]PostedFile

	startupErr error

	responseBuffer *bytes.Buffer
}

// Get is a shortcut for WithVerb("GET") WithPathf(pathFormat, args...)
func (mrb *MockRequestBuilder) Get(pathFormat string, args ...interface{}) *MockRequestBuilder {
	return mrb.WithVerb("GET").WithPathf(pathFormat, args...)
}

// Post is a shortcut for WithVerb("POST") WithPathf(pathFormat, args...)
func (mrb *MockRequestBuilder) Post(pathFormat string, args ...interface{}) *MockRequestBuilder {
	return mrb.WithVerb("POST").WithPathf(pathFormat, args...)
}

// Put is a shortcut for WithVerb("PUT") WithPathf(pathFormat, args...)
func (mrb *MockRequestBuilder) Put(pathFormat string, args ...interface{}) *MockRequestBuilder {
	return mrb.WithVerb("PUT").WithPathf(pathFormat, args...)
}

// Delete is a shortcut for WithVerb("DELETE") WithPathf(pathFormat, args...)
func (mrb *MockRequestBuilder) Delete(pathFormat string, args ...interface{}) *MockRequestBuilder {
	return mrb.WithVerb("DELETE").WithPathf(pathFormat, args...)
}

// WithPathf sets the path for the request.
func (mrb *MockRequestBuilder) WithPathf(pathFormat string, args ...interface{}) *MockRequestBuilder {
	mrb.path = fmt.Sprintf(pathFormat, args...)

	// url.Parse always includes the '/' path prefix.
	if !strings.HasPrefix(mrb.path, "/") {
		mrb.path = fmt.Sprintf("/%s", mrb.path)
	}

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

// WithPostedFile includes a file as a post parameter.
func (mrb *MockRequestBuilder) WithPostedFile(postedFile PostedFile) *MockRequestBuilder {
	if mrb.postedFiles == nil {
		mrb.postedFiles = map[string]PostedFile{}
	}
	mrb.postedFiles[postedFile.Key] = postedFile
	return mrb
}

// WithResponseBuffer optionally sets a response buffer to write to directly.
func (mrb *MockRequestBuilder) WithResponseBuffer(buffer *bytes.Buffer) *MockRequestBuilder {
	mrb.responseBuffer = buffer
	return mrb
}

// Request returns the mock request builder settings as an http.Request.
func (mrb *MockRequestBuilder) Request() (*http.Request, error) {
	req := &http.Request{}

	reqURL, err := url.Parse(fmt.Sprintf("http://localhost%s", mrb.path))

	if err != nil {
		return nil, err
	}

	reqURL.RawQuery = mrb.queryString.Encode()
	req.Method = mrb.verb
	req.URL = reqURL
	req.RequestURI = reqURL.String()
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

	if len(mrb.postedFiles) > 0 {
		b := bytes.NewBuffer(nil)
		w := multipart.NewWriter(b)
		for _, file := range mrb.postedFiles {
			fw, err := w.CreateFormFile(file.Key, file.FileName)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(fw, bytes.NewBuffer(file.Contents))
			if err != nil {
				return nil, err
			}
		}
		// Don't forget to set the content type, this will contain the boundary.
		req.Header.Set("Content-Type", w.FormDataContentType())

		err = w.Close()
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(b)
	} else if len(mrb.postBody) > 0 {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(mrb.postBody))
	}

	return req, nil
}

// Ctx returns the mock request as a request context.
func (mrb *MockRequestBuilder) Ctx(p RouteParameters) (*Ctx, error) {
	r, err := mrb.Request()

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
	var rc *Ctx
	if mrb.app != nil {
		rc = mrb.app.newCtx(w, r, p)
	} else {
		rc = NewCtx(w, r, p)
	}

	return rc, nil
}

// FetchResponse runs the mock request.
func (mrb *MockRequestBuilder) FetchResponse() (res *http.Response, err error) {
	if mrb.startupErr != nil {
		err = mrb.startupErr
		return
	}

	if mrb.app != nil && mrb.app.panicHandler != nil {
		defer func() {
			if r := recover(); r != nil {
				rc, _ := mrb.Ctx(nil)
				controllerResult := mrb.app.panicHandler(rc, r)
				panicRecoveryBuffer := bytes.NewBuffer([]byte{})
				panicRecoveryWriter := NewMockResponseWriter(panicRecoveryBuffer)
				err = controllerResult.Render(NewCtx(panicRecoveryWriter, rc.Request, rc.routeParameters))
				res = &http.Response{
					Body:          ioutil.NopCloser(bytes.NewBuffer(panicRecoveryBuffer.Bytes())),
					ContentLength: int64(panicRecoveryWriter.ContentLength()),
					Header:        http.Header{},
					StatusCode:    panicRecoveryWriter.StatusCode(),
					Proto:         "http",
					ProtoMajor:    1,
					ProtoMinor:    1,
				}
			}
		}()
	}

	handle, params, addTrailingSlash := mrb.app.router.Lookup(mrb.verb, mrb.path)
	if addTrailingSlash {
		mrb.path = mrb.path + "/"
	}

	handle, params, addTrailingSlash = mrb.app.router.Lookup(mrb.verb, mrb.path)
	if handle == nil {
		return nil, exception.Newf("No matching route for path %s `%s`", mrb.verb, mrb.path)
	}

	req, err := mrb.Request()
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
	res = &http.Response{
		Body:          ioutil.NopCloser(bytes.NewBuffer(buffer.Bytes())),
		ContentLength: int64(w.ContentLength()),
		Header:        http.Header{},
		StatusCode:    w.statusCode,
		Proto:         "http",
		ProtoMajor:    1,
		ProtoMinor:    1,
	}

	for key, values := range w.Header() {
		for _, value := range values {
			res.Header.Add(key, value)
		}
	}

	return
}

// FetchResponseAsJSON executes the mock request and reads the response to the given object as json.
func (mrb *MockRequestBuilder) FetchResponseAsJSON(object interface{}) error {
	res, err := mrb.FetchResponse()
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

// FetchResponseAsXML executes the mock request and reads the response to the given object as json.
func (mrb *MockRequestBuilder) FetchResponseAsXML(object interface{}) error {
	res, err := mrb.FetchResponse()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(contents, object)
}

// FetchResponseAsBytes returns the response as bytes.
func (mrb *MockRequestBuilder) FetchResponseAsBytes() ([]byte, error) {
	res, err := mrb.FetchResponse()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// FetchResponseAsBytesWithMeta returns the response as bytes with meta information.
func (mrb *MockRequestBuilder) FetchResponseAsBytesWithMeta() (*ResponseMeta, []byte, error) {
	res, err := mrb.FetchResponse()
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	return NewResponseMeta(res), contents, err
}

// Execute just runs the request.
func (mrb *MockRequestBuilder) Execute() error {
	_, err := mrb.FetchResponseAsBytes()
	return err
}

// ExecuteWithMeta returns basic metadata for a response.
func (mrb *MockRequestBuilder) ExecuteWithMeta() (*ResponseMeta, error) {
	res, err := mrb.FetchResponse()
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errors.New("`res` is nil")
	}

	if res.Body != nil {
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}
	return NewResponseMeta(res), nil
}
