package web

import (
	"fmt"
	"time"

	util "github.com/blendlabs/go-util"
	env "github.com/blendlabs/go-util/env"
)

// NewConfigFromEnv returns a new config from the environment.
func NewConfigFromEnv() *Config {
	var config Config
	env.Env().ReadInto(&config)
	return &config
}

// Config is an object used to set up a web app.
type Config struct {
	Port     int32  `json:"port" yaml:"port" env:"PORT"`
	BindAddr string `json:"bindAddr" yaml:"bindAddr" env:"BIND_ADDR"`
	BaseURL  string `json:"baseURL" yaml:"baseURL" env:"BASE_URL"`

	RedirectTrailingSlash  *bool `json:"redirectTrailingSlash" yaml:"redirectTrailingSlash"`
	HandleOptions          *bool `json:"handleOptions" yaml:"handleOptions"`
	HandleMethodNotAllowed *bool `json:"handleMethodNotAllowed" yaml:"handleMethodNotAllowed"`
	RecoverPanics          *bool `json:"recoverPanics" yaml:"recoverPanics"`

	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes" env:"MAX_HEADER_BYTES"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout" env:"READ_HEADER_TIMEOUT"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout" env:"READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout" env:"WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout" env:"IDLE_TIMEOUT"`

	TLS       TLSConfig         `json:"tls" yaml:"tls"`
	ViewCache ViewCacheConfig   `json:"viewCache" yaml:"viewCache"`
	Auth      AuthManagerConfig `json:"auth" yaml:"auth"`
}

// GetBindAddr coalesces the bind addr, the port, or the default.
func (c Config) GetBindAddr(defaults ...string) string {
	if len(c.BindAddr) > 0 {
		return c.BindAddr
	}
	if c.Port > 0 {
		return fmt.Sprintf(":%d", c.Port)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultBindAddr
}

// GetPort returns the int32 port for a given config.
// This is useful in things like kubernetes pod templates.
// If the config .Port is unset, it will parse the .BindAddr,
// or the DefaultBindAddr for the port number.
func (c Config) GetPort(defaults ...int32) int32 {
	if c.Port > 0 {
		return c.Port
	}
	if len(c.BindAddr) > 0 {
		return PortFromBindAddr(c.BindAddr)
	}
	return PortFromBindAddr(DefaultBindAddr)
}

// GetBaseURL gets a property.
func (c Config) GetBaseURL(defaults ...string) string {
	return util.Coalesce.String(c.BaseURL, "", defaults...)
}

// GetRedirectTrailingSlash returns if we automatically redirect for a missing trailing slash.
func (c Config) GetRedirectTrailingSlash(defaults ...bool) bool {
	return util.Coalesce.Bool(c.RedirectTrailingSlash, DefaultRedirectTrailingSlash, defaults...)
}

// GetHandleOptions returns if we should handle OPTIONS verb requests.
func (c Config) GetHandleOptions(defaults ...bool) bool {
	return util.Coalesce.Bool(c.HandleOptions, DefaultHandleOptions, defaults...)
}

// GetHandleMethodNotAllowed returns if we should handle method not allowed results.
func (c Config) GetHandleMethodNotAllowed(defaults ...bool) bool {
	return util.Coalesce.Bool(c.HandleMethodNotAllowed, DefaultHandleMethodNotAllowed, defaults...)
}

// GetRecoverPanics returns if we should recover panics or not.
func (c Config) GetRecoverPanics(defaults ...bool) bool {
	return util.Coalesce.Bool(c.RecoverPanics, DefaultRecoverPanics, defaults...)
}

// GetMaxHeaderBytes returns the maximum header size in bytes or a default.
func (c Config) GetMaxHeaderBytes(defaults ...int) int {
	return util.Coalesce.Int(c.MaxHeaderBytes, DefaultMaxHeaderBytes, defaults...)
}

// GetReadTimeout gets a property.
func (c Config) GetReadTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.ReadTimeout, DefaultReadTimeout, defaults...)
}

// GetReadHeaderTimeout gets a property.
func (c Config) GetReadHeaderTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.ReadHeaderTimeout, DefaultReadHeaderTimeout, defaults...)
}

// GetWriteTimeout gets a property.
func (c Config) GetWriteTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.WriteTimeout, DefaultWriteTimeout, defaults...)
}

// GetIdleTimeout gets a property.
func (c Config) GetIdleTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.IdleTimeout, DefaultIdleTimeout, defaults...)
}
