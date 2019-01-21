package request

import (
	"net/http"
	"net/url"
	"time"
)

//--------------------------------------------------------------------------------
// RequestMeta
//--------------------------------------------------------------------------------

// NewRequestMeta returns a new meta object for a request.
func NewRequestMeta(req *http.Request) *Meta {
	return &Meta{
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Header,
	}
}

// Meta is a summary of the request meta useful for logging.
type Meta struct {
	// StartTime will be 0 if the request has not been started yet
	StartTime time.Time
	Method    string
	URL       *url.URL
	Headers   http.Header
}
