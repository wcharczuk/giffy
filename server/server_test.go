package server

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New(logger.None(), oauth.Must(oauth.NewFromEnv()), config.NewFromEnv())
	assert.NotNil(app)
}
