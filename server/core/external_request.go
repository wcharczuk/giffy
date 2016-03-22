package core

import (
	"fmt"
	"net/url"

	"github.com/blendlabs/go-request"
)

// NewExternalRequest creates a new external request.
func NewExternalRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().OnResponse(func(meta *request.HTTPResponseMeta, body []byte) {
		fmt.Printf("External Request Response -- %s\n", string(body))
	}).OnRequest(func(verb string, url *url.URL) {
		fmt.Printf("External Request: %s %s\n", verb, url.String())
	})
}
