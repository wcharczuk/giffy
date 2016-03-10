package core

import "github.com/blendlabs/go-request"

// NewExternalRequest creates a new external request.
func NewExternalRequest() *request.HTTPRequest {
	return request.NewHTTPRequest()
}
