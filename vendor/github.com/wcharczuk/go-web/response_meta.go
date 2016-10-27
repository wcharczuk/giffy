package web

import "net/http"

// NewResponseMeta creates a new ResponseMeta.
func NewResponseMeta(res *http.Response) *ResponseMeta {
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
