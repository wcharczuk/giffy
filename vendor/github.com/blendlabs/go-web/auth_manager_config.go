package web

import (
	"time"

	util "github.com/blendlabs/go-util"
)

// AuthManagerConfig is the config for the auth manager.
type AuthManagerConfig struct {
	// UseSessionCache enables or disables the in memory session cache.
	// Note: If the session cache is disabled you *must* provide a fetch handler.
	UseSessionCache *bool `json:"useSessionCache" yaml:"useSessionCache" env:"USE_SESSION_CACHE"`
	// SessionTimeout is a fixed duration to use when calculating hard or rolling deadlines.
	SessionTimeout time.Duration `json:"sessionTimeout" yaml:"sessionTimeout" env:"SESSION_TIMEOUT"`
	// SessionTimeoutIsAbsolute determines if the session timeout is a hard deadline or if it gets pushed forward with usage.
	// The default is to use a hard deadline.
	SessionTimeoutIsAbsolute *bool `json:"sessionTimeoutIsAbsolute" yaml:"sessionTimeoutIsAbsolute" env:"SESSION_TIMEOUT_ABSOLUTE"`
	// CookieHTTPS determines if we should flip the `https only` flag on issued cookies.
	CookieHTTPS *bool `json:"cookieHTTPS" yaml:"cookieHTTPS" env:"COOKIE_HTTPS"`
	// CookieName is the name of the cookie to issue with sessions.
	CookieName string `json:"cookieName" yaml:"cookieName" env:"COOKIE_NAME"`
	// CookiePath is the path on the cookie to issue with sessions.
	CookiePath string `json:"cookiePath" yaml:"cookiePath" env:"COOKIE_PATH"`

	// SecureCookieKey is a key to use to encrypt the sessionID as a second factor cookie.
	SecureCookieKey []byte `json:"secureCookieKey" yaml:"secureCookieKey" env:"SECURE_COOKIE_KEY,base64"`
	// SecureCookieHTTPS determines if we should flip the `https only` flag on issued secure cookies.
	SecureCookieHTTPS *bool `json:"secureCookieHTTPS" yaml:"secureCookieHTTPS" env:"SECURE_COOKIE_HTTPS"`
	// SecureCookieName is the name of the secure cookie to issue with sessions.
	SecureCookieName string `json:"secureCookieName" yaml:"secureCookieName" env:"SECURE_COOKIE_NAME"`
	// SecureCookiePath is the path on the secure cookie to issue with sessions.
	SecureCookiePath string `json:"secureCookiePath,omitempty" yaml:"secureCookiePath,omitempty" env:"SECURE_COOKIE_PATH"`
}

// GetUseSessionCache returns a property or a default.
func (ac AuthManagerConfig) GetUseSessionCache(defaults ...bool) bool {
	return util.Coalesce.Bool(ac.UseSessionCache, DefaultUseSessionCache, defaults...)
}

// GetSessionTimeout returns a property or a default.
func (ac AuthManagerConfig) GetSessionTimeout(defaults ...time.Duration) time.Duration {
	return util.Coalesce.Duration(ac.SessionTimeout, DefaultSessionTimeout, defaults...)
}

// GetSessionTimeoutIsAbsolute returns a property or a default.
func (ac AuthManagerConfig) GetSessionTimeoutIsAbsolute(defaults ...bool) bool {
	return util.Coalesce.Bool(ac.SessionTimeoutIsAbsolute, DefaultSessionTimeoutIsAbsolute, defaults...)
}

// GetCookieHTTPS returns a property or a default.
func (ac AuthManagerConfig) GetCookieHTTPS(defaults ...bool) bool {
	return util.Coalesce.Bool(ac.CookieHTTPS, DefaultCookieHTTPS, defaults...)
}

// GetCookieName returns a property or a default.
func (ac AuthManagerConfig) GetCookieName(defaults ...string) string {
	return util.Coalesce.String(ac.CookieName, DefaultCookieName, defaults...)
}

// GetCookiePath returns a property or a default.
func (ac AuthManagerConfig) GetCookiePath(defaults ...string) string {
	return util.Coalesce.String(ac.CookiePath, DefaultCookiePath, defaults...)
}

// GetSecureCookieKey returns a property or a default.
func (ac AuthManagerConfig) GetSecureCookieKey(defaults ...[]byte) []byte {
	return util.Coalesce.Bytes(ac.SecureCookieKey, nil, defaults...)
}

// GetSecureCookieHTTPS returns a property or a default.
func (ac AuthManagerConfig) GetSecureCookieHTTPS(defaults ...bool) bool {
	return util.Coalesce.Bool(ac.SecureCookieHTTPS, DefaultSecureCookieHTTPS, defaults...)
}

// GetSecureCookieName returns a property or a default.
func (ac AuthManagerConfig) GetSecureCookieName(defaults ...string) string {
	return util.Coalesce.String(ac.SecureCookieName, DefaultSecureCookieName, defaults...)
}

// GetSecureCookiePath returns a property or a default.
func (ac AuthManagerConfig) GetSecureCookiePath(defaults ...string) string {
	return util.Coalesce.String(ac.SecureCookiePath, DefaultCookiePath, defaults...)
}
