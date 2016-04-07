package external

import "github.com/blendlabs/go-request"

// NewRequest creates a new external request.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest()
}
