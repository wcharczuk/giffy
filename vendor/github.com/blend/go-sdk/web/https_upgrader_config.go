package web

import (
	"fmt"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/util"
)

// NewHTTPSUpgraderConfigFromEnv returns an https upgrader config populated from the environment.
func NewHTTPSUpgraderConfigFromEnv() *HTTPSUpgraderConfig {
	var cfg HTTPSUpgraderConfig
	if err := env.Env().ReadInto(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

// HTTPSUpgraderConfig is the config for the https upgrader server.
type HTTPSUpgraderConfig struct {
	Port       int32  `json:"port" yaml:"port" env:"UPGRADE_PORT"`
	BindAddr   string `json:"bindAddr" yaml:"bindAddr" env:"UPGRADE_BIND_ADDR"`
	TargetPort int32  `json:"targetPort" yaml:"targetPort" env:"UPGRADE_TARGET_PORT"`

	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes" env:"MAX_HEADER_BYTES"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout" env:"READ_HEADER_TIMEOUT"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout" env:"READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout" env:"WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout" env:"IDLE_TIMEOUT"`
}

// GetPort returns the int32 port for a given config.
// This is useful in things like kubernetes pod templates.
// If the config .Port is unset, it will parse the .BindAddr,
// or the DefaultBindAddr for the port number.
func (c HTTPSUpgraderConfig) GetPort(defaults ...int32) int32 {
	if c.Port > 0 {
		return c.Port
	}
	if len(c.BindAddr) > 0 {
		return PortFromBindAddr(c.BindAddr)
	}
	return PortFromBindAddr(DefaultBindAddr)
}

// GetBindAddr coalesces the bind addr, the port, or the default.
func (c HTTPSUpgraderConfig) GetBindAddr(defaults ...string) string {
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

// GetTargetPort gets the target port.
// It defaults to unset, i.e. use the https default of 443.
func (c HTTPSUpgraderConfig) GetTargetPort(defaults ...int32) int32 {
	return util.Coalesce.Int32(c.TargetPort, 0, defaults...)
}

// GetMaxHeaderBytes returns the maximum header size in bytes or a default.
func (c HTTPSUpgraderConfig) GetMaxHeaderBytes(defaults ...int) int {
	return util.Coalesce.Int(c.MaxHeaderBytes, DefaultMaxHeaderBytes, defaults...)
}

// GetReadTimeout gets a property.
func (c HTTPSUpgraderConfig) GetReadTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.ReadTimeout, DefaultReadTimeout, defaults...)
}

// GetReadHeaderTimeout gets a property.
func (c HTTPSUpgraderConfig) GetReadHeaderTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.ReadHeaderTimeout, DefaultReadHeaderTimeout, defaults...)
}

// GetWriteTimeout gets a property.
func (c HTTPSUpgraderConfig) GetWriteTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.WriteTimeout, DefaultWriteTimeout, defaults...)
}

// GetIdleTimeout gets a property.
func (c HTTPSUpgraderConfig) GetIdleTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.IdleTimeout, DefaultIdleTimeout, defaults...)
}
