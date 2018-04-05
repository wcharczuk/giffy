package oauth

import (
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/util"
)

const (
	// DefaultNonceTimeout is the default timeout before nonces are no longer honored.
	DefaultNonceTimeout = 3 * time.Hour
)

// NewConfigFromEnv creates a new config from the environment.
func NewConfigFromEnv() *Config {
	var cfg Config
	err := env.Env().ReadInto(&cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}

// Config is the config options.
type Config struct {
	Secret string `json:"secret" yaml:"secret" env:"GOOGLE_SECRET"`

	SkipDomainValidation bool     `json:"skipDomainValidation" yaml:"skipDomainValidation" env:"GOOGLE_SKIP_DOMAIN_VALIDATION"`
	RedirectURI          string   `json:"redirectURI" yaml:"redirectURI" env:"GOOGLE_REDIRECT_URI"`
	ValidDomains         []string `json:"validDomains" yaml:"validDomains" env:"GOOGLE_VALID_DOMAINS,csv"`
	HostedDomain         string   `json:"hostedDomain" yaml:"hostedDomain" env:"GOOGLE_HOSTED_DOMAIN"`

	ClientID     string `json:"clientID" yaml:"clientID" env:"GOOGLE_CLIENT_ID"`
	ClientSecret string `json:"clientSecret" yaml:"clientSecret" env:"GOOGLE_CLIENT_SECRET"`

	NonceTimeout time.Duration `json:"nonceTimeout" yaml:"nonceTimeout" env:"GOOGLE_NONCE_TIMEOUT"`
}

// IsZero returns if the config is set or not.
func (c Config) IsZero() bool {
	return len(c.ClientID) == 0 || len(c.ClientSecret) == 0
}

// GetSecret gets the secret if set or a default.
func (c Config) GetSecret(defaults ...[]byte) []byte {
	if len(c.Secret) > 0 {
		decoded, _ := Base64Decode(c.Secret)
		return decoded
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetSkipDomainValidation returns if we should skip domain validation.
func (c Config) GetSkipDomainValidation() bool {
	return c.SkipDomainValidation
}

// GetRedirectURI returns a property or a default.
func (c Config) GetRedirectURI(inherited ...string) string {
	return util.Coalesce.String(c.RedirectURI, "", inherited...)
}

// GetValidDomains returns a property or a default.
func (c Config) GetValidDomains(inherited ...[]string) []string {
	return util.Coalesce.Strings(c.ValidDomains, nil, inherited...)
}

// GetHostedDomain returns a property or a default.
func (c Config) GetHostedDomain(inherited ...string) string {
	return util.Coalesce.String(c.HostedDomain, "", inherited...)
}

// GetClientID returns a property or a default.
func (c Config) GetClientID(inherited ...string) string {
	return util.Coalesce.String(c.ClientID, "", inherited...)
}

// GetClientSecret returns a property or a default.
func (c Config) GetClientSecret(inherited ...string) string {
	return util.Coalesce.String(c.ClientSecret, "", inherited...)
}

// GetNonceTimeout returns the nonce timeout or a default.
func (c Config) GetNonceTimeout(inherited ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.NonceTimeout, DefaultNonceTimeout, inherited...)
}
