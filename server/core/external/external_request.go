package external

import (
	"net/url"

	"github.com/wcharczuk/go-web"

	"github.com/blendlabs/go-request"
	"github.com/wcharczuk/giffy/server/core"
)

// NewRequest creates a new external request.
func NewRequest() *request.HTTPRequest {
	if core.ConfigEnvironment() == "prod" {
		return request.NewHTTPRequest()
	}

	return request.NewHTTPRequest().OnResponse(func(meta *request.HTTPResponseMeta, body []byte) {
		web.Logf("External Request Response -- %s\n", string(body))
	}).OnRequest(func(verb string, url *url.URL) {
		web.Logf("External Request Outgoing -- %s %s\n", verb, url.String())
	})
}
