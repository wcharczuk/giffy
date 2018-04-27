package cron

import (
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/util"
)

// NewConfigFromEnv creates a new config from the environment.
func NewConfigFromEnv() *Config {
	var cfg Config
	if err := env.Env().ReadInto(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

// Config is the config object.
type Config struct {
	HeartbeatInterval time.Duration `json:"heartbeatInterval" yaml:"heartbeatInterval" env:"CRON_HEARTBEAT_INTERVAL"`
}

// GetHeartbeatInterval gets a property or a default.
func (c Config) GetHeartbeatInterval(inherited ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.HeartbeatInterval, DefaultHeartbeatInterval, inherited...)
}
