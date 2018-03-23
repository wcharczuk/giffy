package spiffy

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestDefault(t *testing.T) {
	assert := assert.New(t)

	assert.NotNil(Default())
}
