package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/giffy/server/config"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New(config.NewFromEnv())
	assert.NotNil(app)
}
