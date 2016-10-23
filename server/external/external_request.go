package external

import (
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-request"
)

// NewRequest creates a new external request.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().WithDiagnostics(logger.Diagnostics())
}
