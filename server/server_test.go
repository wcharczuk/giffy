package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
	logger "github.com/blendlabs/go-logger"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New(logger.None(), config.NewFromEnv())
	assert.NotNil(app)
}
