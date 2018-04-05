package configutil

import (
	"bytes"
	"testing"

	assert "github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util/env"
)

type Config struct {
	Environment string `json:"env" yaml:"env" env:"SERVICE_ENV"`
	Other       string `json:"other" yaml:"other" env:"OTHER"`
}

func TestRead(t *testing.T) {
	assert := assert.New(t)
	defer env.Restore()

	env.Env().Set(env.VarServiceEnv, "dev")
	var cfg Config
	err := ReadFromReader(&cfg, bytes.NewBuffer([]byte("env: test\nother: foo")), ExtensionYAML)
	assert.Nil(err)
	assert.Equal("dev", cfg.Environment)
}

func TestReadPathUnset(t *testing.T) {
	assert := assert.New(t)
	defer env.Restore()

	env.Env().Set(env.VarServiceEnv, "dev")
	var cfg Config
	err := ReadFromPath(&cfg, "")
	assert.True(IsPathUnset(err))
	assert.Equal("dev", cfg.Environment)
}
