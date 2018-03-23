package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
	google "github.com/blendlabs/go-google-oauth"
	logger "github.com/blendlabs/go-logger"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New(logger.None(), google.NewFromEnv(), config.NewFromEnv())
	assert.NotNil(app)
}
