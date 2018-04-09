package web

import (
	"fmt"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/util"
)

// NewHealthzConfigFromEnv returns a new config from the environment.
func NewHealthzConfigFromEnv() *HealthzConfig {
	var cfg HealthzConfig
	if err := env.Env().ReadInto(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

// HealthzConfig is an object used to set up a healthz sidecar.
type HealthzConfig struct {
	Port     int32  `json:"port" yaml:"port" env:"HEALTHZ_PORT"`
	BindAddr string `json:"bindAddr" yaml:"bindAddr" env:"HEALTHZ_BIND_ADDR"`

	// DefaultHeaders are included on any responses. The app ships with a set of default headers, which you can augment with this property.
	DefaultHeaders map[string]string `json:"defaultHeaders" yaml:"defaultHeaders"`

	RecoverPanics     *bool         `json:"recoverPanics" yaml:"recoverPanics"`
	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes" env:"HEALTHZ_MAX_HEADER_BYTES"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout" env:"HEALTHZ_READ_HEADER_TIMEOUT"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout" env:"HEALTHZ_READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout" env:"HEALTHZ_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout" env:"HEALTHZ_IDLE_TIMEOUT"`
}

// GetBindAddr util.Coalesces the bind addr, the port, or the default.
func (hc HealthzConfig) GetBindAddr(defaults ...string) string {
	if len(hc.BindAddr) > 0 {
		return hc.BindAddr
	}
	if hc.Port > 0 {
		return fmt.Sprintf(":%d", hc.Port)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultHealthzBindAddr
}

// GetPort returns the int32 port for a given config.
// This is useful in things like kubernetes pod templates.
// If the config .Port is unset, it will parse the .BindAddr,
// or the DefaultBindAddr for the port number.
func (hc HealthzConfig) GetPort(defaults ...int32) int32 {
	if hc.Port > 0 {
		return hc.Port
	}
	if len(hc.BindAddr) > 0 {
		return PortFromBindAddr(hc.BindAddr)
	}
	return PortFromBindAddr(DefaultHealthzBindAddr)
}

// GetRecoverPanics returns if we should recover panics or not.
func (hc HealthzConfig) GetRecoverPanics(defaults ...bool) bool {
	return util.Coalesce.Bool(hc.RecoverPanics, DefaultRecoverPanics, defaults...)
}

// GetMaxHeaderBytes returns the maximum header size in bytes or a default.
func (hc HealthzConfig) GetMaxHeaderBytes(defaults ...int) int {
	return util.Coalesce.Int(hc.MaxHeaderBytes, DefaultMaxHeaderBytes, defaults...)
}

// GetReadTimeout gets a property.
func (hc HealthzConfig) GetReadTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(hc.ReadTimeout, DefaultReadTimeout, defaults...)
}

// GetReadHeaderTimeout gets a property.
func (hc HealthzConfig) GetReadHeaderTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(hc.ReadHeaderTimeout, DefaultReadHeaderTimeout, defaults...)
}

// GetWriteTimeout gets a property.
func (hc HealthzConfig) GetWriteTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(hc.WriteTimeout, DefaultWriteTimeout, defaults...)
}

// GetIdleTimeout gets a property.
func (hc HealthzConfig) GetIdleTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(hc.IdleTimeout, DefaultIdleTimeout, defaults...)
}
