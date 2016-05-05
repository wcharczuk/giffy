package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)
	app := Init()
	assert.NotNil(app)
}
