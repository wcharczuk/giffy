package external

import (
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/request"
)

// NewRequest creates a new external request.
func NewRequest() *request.Request {
	return request.New().WithLogger(logger.Default())
}
