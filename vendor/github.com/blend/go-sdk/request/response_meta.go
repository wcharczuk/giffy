package request

import (
	"net/http"
	"strings"
	"time"
)

// NewResponseMeta returns a new meta object for a response.
func NewResponseMeta(res *http.Response) *ResponseMeta {
	if res == nil {
		return nil
	}
	return &ResponseMeta{
		CompleteTime:    now(),
		StatusCode:      res.StatusCode,
		ContentLength:   res.ContentLength,
		ContentType:     tryHeader(res.Header, "Content-Type", "content-type"),
		ContentEncoding: tryHeader(res.Header, "Content-Encoding", "content-encoding"),
		Headers:         res.Header,
		Cert:            NewCertInfo(res),
	}
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

func tryHeader(headers http.Header, keys ...string) string {
	for _, key := range keys {
		if values, hasValues := headers[key]; hasValues {
			return strings.Join(values, ";")
		}
	}
	return ""
}
