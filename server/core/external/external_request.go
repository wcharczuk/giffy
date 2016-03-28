package external

import (
	"net/url"

	"github.com/blendlabs/go-request"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/web"
)

// NewRequest creates a new external request.
func NewRequest() *request.HTTPRequest {
	if core.ConfigEnvironment() == "prod" {
		return request.NewHTTPRequest()
	}

	return request.NewHTTPRequest().OnResponse(func(meta *request.HTTPResponseMeta, body []byte) {
		web.Logf("External Request Response -- %s\n", string(body))
	}).OnRequest(func(verb string, url *url.URL) {
		web.Logf("External Request -- %s %s\n", verb, url.String())
	})
}
