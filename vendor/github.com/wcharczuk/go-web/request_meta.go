package web

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// NewRequestMeta returns a new meta object for a request.
func NewRequestMeta(req *http.Request) *RequestMeta {
	return &RequestMeta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
	}
}

// NewRequestMetaWithBody returns a new meta object for a request and reads the body.
func NewRequestMetaWithBody(req *http.Request) (*RequestMeta, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return &RequestMeta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
		Body:    body,
	}, nil
}

// RequestMeta is the metadata for a request.
type RequestMeta struct {
	StartTime time.Time
	Verb      string
	URL       *url.URL
	Headers   http.Header
	Body      []byte
}
