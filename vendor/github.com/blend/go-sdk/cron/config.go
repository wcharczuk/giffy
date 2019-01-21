package cron

import (
	"github.com/blend/go-sdk/env"
)

// NewConfigFromEnv creates a new config from the environment.
func NewConfigFromEnv() (*Config, error) {
	var cfg Config
	if err := env.Env().ReadInto(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// MustNewConfigFromEnv returns a new config set from environment variables,
// it will panic if there is an error.
func MustNewConfigFromEnv() *Config {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		panic(err)
	}
	return cfg
}

// Config is the config object.
type Config struct{}
