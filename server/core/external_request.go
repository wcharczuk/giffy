package core

import (
	"fmt"
	"net/url"
	"time"

	"github.com/blendlabs/go-request"
)

// NewExternalRequest creates a new external request.
func NewExternalRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().
		OnRequest(func(verb string, url *url.URL) {
			fmt.Printf("%s - Outgoing Request - %s %s\n", time.Now().UTC().Format(time.RFC3339), verb, url.String())
		}).
		OnResponse(func(meta *request.HTTPResponseMeta, body []byte) {
			println(string(body))
		})
}
