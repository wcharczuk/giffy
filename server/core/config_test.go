package core

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestConfigLocalIP(t *testing.T) {
	assert := assert.New(t)
	localIP := ConfigLocalIP()
	assert.NotEmpty(localIP)
}

func TestConfigKey(t *testing.T) {
	assert := assert.New(t)
	key := ConfigKey()
	assert.NotEmpty(key)
}
