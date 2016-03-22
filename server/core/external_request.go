package core

import (
	"fmt"

	"github.com/blendlabs/go-request"
)

// NewExternalRequest creates a new external request.
func NewExternalRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().OnResponse(func(meta *request.HTTPResponseMeta, body []byte) {
		fmt.Printf("External Response -- %s\n", string(body))
	})
}
