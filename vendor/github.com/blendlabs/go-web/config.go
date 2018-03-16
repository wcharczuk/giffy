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

	// HSTS determines if we should issue the Strict-Transport-Security header.
	HSTS                  *bool `json:"hsts" yaml:"hsts"`
	HSTSMaxAgeSeconds     int   `json:"hstsMaxAgeSeconds" yaml:"hstsMaxAgeSeconds"`
	HSTSIncludeSubDomains *bool `json:"hstsIncludeSubdomains" yaml:"hstsIncludeSubdomains"`
	HSTSPreload           *bool `json:"hstsPreload" yaml:"hstsPreload"`

	// UseSessionCache enables or disables the in memory session cache.
	// Note: If the session cache is disabled you *must* provide a fetch handler.
	UseSessionCache *bool `json:"useSessionCache" yaml:"useSessionCache" env:"USE_SESSION_CACHE"`
	// SessionTimeout is a fixed duration to use when calculating hard or rolling deadlines.
	SessionTimeout time.Duration `json:"sessionTimeout" yaml:"sessionTimeout" env:"SESSION_TIMEOUT"`
	// SessionTimeoutIsAbsolute determines if the session timeout is a hard deadline or if it gets pushed forward with usage.
	// The default is to use a hard deadline.
	SessionTimeoutIsAbsolute *bool `json:"sessionTimeoutIsAbsolute" yaml:"sessionTimeoutIsAbsolute" env:"SESSION_TIMEOUT_ABSOLUTE"`
	// CookieHTTPS determines if we should flip the `https only` flag on issued cookies.
	CookieHTTPSOnly *bool `json:"cookieHTTPSOnly" yaml:"cookieHTTPSOnly" env:"COOKIE_HTTPS_ONLY"`
	// CookieName is the name of the cookie to issue with sessions.
	CookieName string `json:"cookieName" yaml:"cookieName" env:"COOKIE_NAME"`
	// CookiePath is the path on the cookie to issue with sessions.
	CookiePath string `json:"cookiePath" yaml:"cookiePath" env:"COOKIE_PATH"`

	// SecureCookieKey is a key to use to encrypt the sessionID as a second factor cookie.
	SecureCookieKey []byte `json:"secureCookieKey" yaml:"secureCookieKey" env:"SECURE_COOKIE_KEY,base64"`
	// SecureCookieHTTPS determines if we should flip the `https only` flag on issued secure cookies.
	SecureCookieHTTPSOnly *bool `json:"secureCookieHTTPSOnly" yaml:"secureCookieHTTPSOnly" env:"SECURE_COOKIE_HTTPS_ONLY"`
	// SecureCookieName is the name of the secure cookie to issue with sessions.
	SecureCookieName string `json:"secureCookieName" yaml:"secureCookieName" env:"SECURE_COOKIE_NAME"`
	// SecureCookiePath is the path on the secure cookie to issue with sessions.
	SecureCookiePath string `json:"secureCookiePath,omitempty" yaml:"secureCookiePath,omitempty" env:"SECURE_COOKIE_PATH"`

	// DefaultHeaders are included on any responses. The app ships with a set of default headers, which you can augment with this property.
	DefaultHeaders map[string]string `json:"defaultHeaders" yaml:"defaultHeaders"`

	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes" env:"MAX_HEADER_BYTES"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout" env:"READ_HEADER_TIMEOUT"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout" env:"READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout" env:"WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout" env:"IDLE_TIMEOUT"`

	TLS   TLSConfig       `json:"tls" yaml:"tls"`
	Views ViewCacheConfig `json:"views" yaml:"views"`
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

// GetDefaultHeaders returns the default headers from the config.
func (c Config) GetDefaultHeaders(inherited ...map[string]string) map[string]string {
	output := map[string]string{}
	if len(inherited) > 0 {
		for _, set := range inherited {
			for key, value := range set {
				output[key] = value
			}
		}
	}
	for key, value := range c.DefaultHeaders {
		output[key] = value
	}
	return output
}

// GetHSTS returns a property or a default.
func (c Config) GetHSTS(inherited ...bool) bool {
	return util.Coalesce.Bool(c.HSTS, DefaultHSTS && (c.ListenTLS() || c.BaseURLIsSecureScheme()), inherited...)
}

// GetHSTSMaxAgeSeconds returns a property or a default.
func (c Config) GetHSTSMaxAgeSeconds(inherited ...int) int {
	return util.Coalesce.Int(c.HSTSMaxAgeSeconds, DefaultHSTSMaxAgeSeconds, inherited...)
}

// GetHSTSIncludeSubDomains returns a property or a default.
func (c Config) GetHSTSIncludeSubDomains(inherited ...bool) bool {
	return util.Coalesce.Bool(c.HSTSIncludeSubDomains, DefaultHSTSIncludeSubdomains, inherited...)
}

// GetHSTSPreload returns a property or a default.
func (c Config) GetHSTSPreload(inherited ...bool) bool {
	return util.Coalesce.Bool(c.HSTSPreload, DefaultHSTSPreload, inherited...)
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

// GetUseSessionCache returns a property or a default.
func (c Config) GetUseSessionCache(defaults ...bool) bool {
	return util.Coalesce.Bool(c.UseSessionCache, DefaultUseSessionCache, defaults...)
}

// GetSessionTimeout returns a property or a default.
func (c Config) GetSessionTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.SessionTimeout, DefaultSessionTimeout, defaults...)
}

// GetSessionTimeoutIsAbsolute returns a property or a default.
func (c Config) GetSessionTimeoutIsAbsolute(defaults ...bool) bool {
	return util.Coalesce.Bool(c.SessionTimeoutIsAbsolute, DefaultSessionTimeoutIsAbsolute, defaults...)
}

// ListenTLS returns if we should use a tls listener.
func (c Config) ListenTLS() bool {
	return c.TLS.HasKeyPair()
}

// BaseURLIsSecureScheme returns if the base url starts with https.
func (c Config) BaseURLIsSecureScheme() bool {
	baseURL := c.GetBaseURL()
	if len(baseURL) == 0 {
		return false
	}
	return util.String.HasPrefixCaseInsensitive(baseURL, "https://") || util.String.HasPrefixCaseInsensitive(baseURL, "spdy://")
}

// GetCookieHTTPSOnly returns a property or a default.
func (c Config) GetCookieHTTPSOnly(defaults ...bool) bool {
	return util.Coalesce.Bool(c.CookieHTTPSOnly, c.ListenTLS() || c.BaseURLIsSecureScheme(), defaults...)
}

// GetCookieName returns a property or a default.
func (c Config) GetCookieName(defaults ...string) string {
	return util.Coalesce.String(c.CookieName, DefaultCookieName, defaults...)
}

// GetCookiePath returns a property or a default.
func (c Config) GetCookiePath(defaults ...string) string {
	return util.Coalesce.String(c.CookiePath, DefaultCookiePath, defaults...)
}

// GetSecureCookieKey returns a property or a default.
func (c Config) GetSecureCookieKey(defaults ...[]byte) []byte {
	return util.Coalesce.Bytes(c.SecureCookieKey, nil, defaults...)
}

// GetSecureCookieHTTPSOnly returns a property or a default.
func (c Config) GetSecureCookieHTTPSOnly(defaults ...bool) bool {
	return util.Coalesce.Bool(c.SecureCookieHTTPSOnly, c.GetCookieHTTPSOnly(), defaults...)
}

// GetSecureCookieName returns a property or a default.
func (c Config) GetSecureCookieName(defaults ...string) string {
	return util.Coalesce.String(c.SecureCookieName, DefaultSecureCookieName, defaults...)
}

// GetSecureCookiePath returns a property or a default.
func (c Config) GetSecureCookiePath(defaults ...string) string {
	return util.Coalesce.String(c.SecureCookiePath, DefaultCookiePath, defaults...)
}
