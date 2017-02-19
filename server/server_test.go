package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	app := New()
	assert.NotNil(app)
}
