package web

import "net/http"

// RawResult is for when you just want to dump bytes.
type RawResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// Render renders the result.
func (rr *RawResult) Render(rc *RequestContext) error {
	if len(rr.ContentType) != 0 {
		rc.Response.Header().Set("Content-Type", rr.ContentType)
	}
	if rr.StatusCode == 0 {
		rc.Response.WriteHeader(http.StatusOK)
	} else {
		rc.Response.WriteHeader(rr.StatusCode)
	}
	_, err := rc.Response.Write(rr.Body)
	return err
}
