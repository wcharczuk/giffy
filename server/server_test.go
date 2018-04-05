package server

import (
	"testing"

	logger "github.com/blend/go-sdk/logger"
	"github.com/blendlabs/go-assert"
	google "github.com/blendlabs/go-google-oauth"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New(logger.None(), google.NewFromEnv(), config.NewFromEnv())
	assert.NotNil(app)
}
