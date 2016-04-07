package external

import (
	"fmt"

	"github.com/blendlabs/go-request"
)

// NewRequest creates a new external request.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().OnResponse(func(meta *request.HTTPResponseMeta, content []byte) {
		fmt.Printf("Response: %s\n", string(content))
	})
}
