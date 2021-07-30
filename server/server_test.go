package server

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app, err := New(config.MustNewFromEnv())
	assert.Nil(err)
	assert.NotNil(app)
}
