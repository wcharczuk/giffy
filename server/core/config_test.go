package core

import (
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestConfigLocalIP(t *testing.T) {
	assert := assert.New(t)
	localIP := ConfigLocalIP()
	assert.NotEmpty(localIP)
}
