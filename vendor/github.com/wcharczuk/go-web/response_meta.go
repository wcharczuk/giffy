package web

import "net/http"

// NewResponseMetaFromResponse creates a new ResponseMeta.
func NewResponseMetaFromResponse(res *http.Response) *ResponseMeta {
	return &ResponseMeta{
		StatusCode:    res.StatusCode,
		Headers:       res.Header,
		ContentLength: res.ContentLength,
	}
}

// ResponseMeta is a metadata response struct
type ResponseMeta struct {
	StatusCode    int
	ContentLength int64
	Headers       http.Header
}
