package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//--------------------------------------------------------------------------------
// RequestMeta
//--------------------------------------------------------------------------------

// NewMeta returns a new meta object for a request.
func NewMeta(req *http.Request) *Meta {
	return &Meta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
	}
}

// NewMetaWithBody returns a new meta object for a request and reads the body.
func NewMetaWithBody(req *http.Request) (*Meta, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return &Meta{
		Verb:    req.Method,
		URL:     req.URL,
		Headers: req.Header,
		Body:    body,
	}, nil
}

// Meta is a summary of the request meta useful for logging.
type Meta struct {
	StartTime time.Time
	Verb      string
	URL       *url.URL
	Headers   http.Header
	Body      []byte
}

//--------------------------------------------------------------------------------
// HttpResponseMeta
//--------------------------------------------------------------------------------

// NewResponseMeta returns a new meta object for a response.
func NewResponseMeta(res *http.Response) *ResponseMeta {
	meta := &ResponseMeta{}

	if res == nil {
		return meta
	}

	meta.CompleteTime = time.Now().UTC()
	meta.StatusCode = res.StatusCode
	meta.ContentLength = res.ContentLength

	meta.ContentType = tryHeader(res.Header, "Content-Type", "content-type")
	meta.ContentEncoding = tryHeader(res.Header, "Content-Encoding", "content-encoding")

	meta.Headers = res.Header
	meta.Cert = NewCertInfo(res)
	return meta
}

func tryHeader(headers http.Header, keys ...string) string {
	for _, key := range keys {
		if values, hasValues := headers[key]; hasValues {
			return strings.Join(values, ";")
		}
	}
	return ""
}

// ResponseMeta is just the meta information for an http response.
type ResponseMeta struct {
	Cert            *CertInfo
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
type ResponseHandler func(req *Meta, meta *ResponseMeta, content []byte)

// StatefulResponseHandler is a receiver for `OnResponse` that includes a state object.
type StatefulResponseHandler func(req *Meta, res *ResponseMeta, content []byte, state interface{})

// OutgoingRequestHandler is a receiver for `OnRequest`.
type OutgoingRequestHandler func(req *Meta)

// MockedResponseProvider is a mocked response provider.
type MockedResponseProvider func(*Request) *MockedResponse

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

// NewCertInfo returns a new cert info from a response.
func NewCertInfo(res *http.Response) *CertInfo {
	if res.TLS != nil && len(res.TLS.PeerCertificates) > 0 {
		var earliestExpiration time.Time
		var latestNotBefore time.Time
		for _, cert := range res.TLS.PeerCertificates {
			if earliestExpiration.IsZero() || earliestExpiration.After(cert.NotAfter) {
				earliestExpiration = cert.NotAfter
			}
			if latestNotBefore.IsZero() || latestNotBefore.Before(cert.NotBefore) {
				latestNotBefore = cert.NotBefore
			}
		}

		firstCert := res.TLS.PeerCertificates[0]

		var issuerCommonName string
		if len(firstCert.Issuer.CommonName) > 0 {
			issuerCommonName = firstCert.Issuer.CommonName
		} else {
			for _, name := range firstCert.Issuer.Names {
				if name.Type.String() == "2.5.4.3" {
					issuerCommonName = fmt.Sprintf("%v", name.Value)
				}
			}
		}

		return &CertInfo{
			DNSNames:         firstCert.DNSNames,
			NotAfter:         earliestExpiration,
			NotBefore:        latestNotBefore,
			IssuerCommonName: issuerCommonName,
		}
	}

	return nil
}

// CertInfo is the information for a certificate.
type CertInfo struct {
	IssuerCommonName string
	DNSNames         []string
	NotAfter         time.Time
	NotBefore        time.Time
}
