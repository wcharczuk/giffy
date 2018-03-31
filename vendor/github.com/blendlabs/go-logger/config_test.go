package logger

import (
	"strings"
	"testing"

	assert "github.com/blendlabs/go-assert"
	env "github.com/blendlabs/go-util/env"
)

func TestNewConfigFlags(t *testing.T) {
	assert := assert.New(t)

	cfg := &Config{Flags: []string{"all", "-debug"}}
	log := NewFromConfig(cfg)
	defer log.Close()

	assert.True(log.IsEnabled(Silly))
	assert.True(log.IsEnabled(Info))
	assert.False(log.IsEnabled(Debug))
	assert.True(log.IsEnabled(Warning))
	assert.True(log.IsEnabled(Error))
	assert.True(log.IsEnabled(Fatal))
	assert.True(log.IsEnabled(Audit))
	assert.True(log.IsEnabled(WebRequest))
}

func TestNewConfigFromEnv(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfigFromEnv()

	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "web.request"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "info"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "warning"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "error"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "fatal"
	})
	assert.None(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "debug"
	})
}

func TestNewConfigFromEnvWithVars(t *testing.T) {
	assert := assert.New(t)

	env.SetEnv(env.Vars{
		"LOG_EVENTS": "info,debug,error,test",
		"LOG_HIDDEN": "debug",
	})
	defer env.Restore()

	cfg := NewConfigFromEnv()

	assert.NotEmpty(cfg.GetFlags())
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "info"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "debug"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "error"
	})
	assert.Any(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "test"
	})
	assert.None(cfg.GetFlags(), func(v Any) bool {
		return v.(string) == "fatal"
	})

}

func TestGetWritersWithOutputFormat(t *testing.T) {
	assert := assert.New(t)

	config := &Config{OutputFormat: string(OutputFormatJSON)}
	writers := config.GetWriters()
	assert.Len(writers, 1)
	assert.Equal(OutputFormatJSON, writers[0].OutputFormat())
	config.OutputFormat = string(OutputFormatText)
	writers = config.GetWriters()
	assert.Len(writers, 1)
	assert.Equal(OutputFormatText, writers[0].OutputFormat())
	config.OutputFormat = "nope"
	writers = config.GetWriters()
	assert.Len(writers, 1)
	assert.Equal(OutputFormatText, writers[0].OutputFormat())
	config.OutputFormat = strings.ToUpper(string(OutputFormatJSON))
	writers = config.GetWriters()
	assert.Len(writers, 1)
	assert.Equal(OutputFormatJSON, writers[0].OutputFormat())
}
