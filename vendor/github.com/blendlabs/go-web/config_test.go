package web

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
	env "github.com/blendlabs/go-util/env"
)

func TestNewConfigFromEnv(t *testing.T) {
	assert := assert.New(t)
	defer env.Restore()

	env.Env().Set("AUTH_SECRET", Base64Encode(util.Crypto.MustCreateKey(32)))

	config := NewConfigFromEnv()
	assert.NotEmpty(config.GetAuthSecret())
}
