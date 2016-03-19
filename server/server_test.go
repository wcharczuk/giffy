package server

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)
	r := Init()
	assert.NotNil(r)
}
