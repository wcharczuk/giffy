package db

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/util"
	"github.com/lib/pq"
)

const (
	// EnvVarDatabaseURL is an environment variable.
	EnvVarDatabaseURL = "DATABASE_URL"

	// DefaultHost is the default database hostname, typically used
	// when developing locally.
	DefaultHost = "localhost"
	// DefaultPort is the default postgres port.
	DefaultPort = "5432"
	// DefaultDatabase is the default database to connect to, we use
	// `postgres` to not pollute the template databases.
	DefaultDatabase = "postgres"
	// DefaultSSLMode is the default connection ssl mode.
	// We use `disable` because typical dev installs do not provision certs.
	DefaultSSLMode = SSLModeDisable

	// SSLModeDisable is an ssl mode.
	// Postgres Docs: "I don't care about security, and I don't want to pay the overhead of encryption."
	SSLModeDisable = "disable"
	// SSLModeAllow is an ssl mode.
	// Postgres Docs: "I don't care about security, but I will pay the overhead of encryption if the server insists on it."
	SSLModeAllow = "allow"
	// SSLModePrefer is an ssl mode.
	// Postgres Docs: "I don't care about encryption, but I wish to pay the overhead of encryption if the server supports it"
	SSLModePrefer = "prefer"
	// SSLModeRequire is an ssl mode.
	// Postgres Docs: "I want my data to be encrypted, and I accept the overhead. I trust that the network will make sure I always connect to the server I want."
	SSLModeRequire = "require"
	// SSLModeVerifyCA is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server that I trust."
	SSLModeVerifyCA = "verify-ca"
	// SSLModeVerifyFull is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server I trust, and that it's the one I specify."
	SSLModeVerifyFull = "verify-full"

	// DefaultUseStatementCache is the default if we should enable the statement cache.
	DefaultUseStatementCache = true
	// DefaultIdleConnections is the default number of idle connections.
	DefaultIdleConnections = 16
	// DefaultMaxConnections is the default maximum number of connections.
	DefaultMaxConnections = 32
	// DefaultMaxLifetime is the default maximum lifetime of driver connections.
	DefaultMaxLifetime time.Duration = 0
	// DefaultBufferPoolSize is the default number of buffer pool entries to maintain.
	DefaultBufferPoolSize = 1024
)

// NewConfig creates a new config.
func NewConfig() *Config {
	return &Config{}
}

// NewConfigFromDSN creates a new config from a dsn.
func NewConfigFromDSN(dsn string) (*Config, error) {
	parsed, err := pq.ParseURL(dsn)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	var config Config
	pieces := util.String.SplitOnSpace(parsed)
	for _, piece := range pieces {
		if strings.HasPrefix(piece, "host=") {
			config.Host = strings.TrimPrefix(piece, "host=")
		} else if strings.HasPrefix(piece, "port=") {
			config.Port = strings.TrimPrefix(piece, "port=")
		} else if strings.HasPrefix(piece, "dbname=") {
			config.Database = strings.TrimPrefix(piece, "dbname=")
		} else if strings.HasPrefix(piece, "user=") {
			config.Username = strings.TrimPrefix(piece, "user=")
		} else if strings.HasPrefix(piece, "password=") {
			config.Password = strings.TrimPrefix(piece, "password=")
		} else if strings.HasPrefix(piece, "sslmode=") {
			config.SSLMode = strings.TrimPrefix(piece, "sslmode=")
		}
	}
	return &config, nil
}

// NewConfigFromEnv returns a new config from the environment.
// The environment variable mappings are as follows:
//	-	DATABSE_URL 	= DSN 	//note that this has precedence over other vars (!!)
// 	-	DB_HOST 		= Host
//	-	DB_PORT 		= Port
//	- 	DB_NAME 		= Database
//	-	DB_SCHEMA		= Schema
//	-	DB_USER 		= Username
//	-	DB_PASSWORD 	= Password
//	-	DB_SSLMODE 		= SSLMode
func NewConfigFromEnv() *Config {
	var config Config
	env.Env().ReadInto(&config)
	return &config
}

// Config is a set of connection config options.
type Config struct {
	// DSN is a fully formed DSN (this skips DSN formation from all other variables outside `schema`).
	DSN string `json:"dsn" yaml:"dsn" env:"DATABASE_URL"`
	// Host is the server to connect to.
	Host string `json:"host" yaml:"host" env:"DB_HOST"`
	// Port is the port to connect to.
	Port string `json:"port" yaml:"port" env:"DB_PORT"`
	// DBName is the database name
	Database string `json:"database" yaml:"database" env:"DB_NAME"`
	// Schema is the application schema within the database, defaults to `public`.
	Schema string `json:"schema" yaml:"schema" env:"DB_SCHEMA"`
	// Username is the username for the connection via password auth.
	Username string `json:"username" yaml:"username" env:"DB_USER"`
	// Password is the password for the connection via password auth.
	Password string `json:"password" yaml:"password" env:"DB_PASSWORD"`
	// SSLMode is the sslmode for the connection.
	SSLMode string `json:"sslMode" yaml:"sslMode" env:"DB_SSLMODE"`
	// UseStatementCache indicates if we should use the prepared statement cache.
	UseStatementCache *bool `json:"useStatementCache" yaml:"useStatementCache" env:"DB_USE_STATEMENT_CACHE"`
	// IdleConnections is the number of idle connections.
	IdleConnections int `json:"idleConnections" yaml:"idleConnections" env:"DB_IDLE_CONNECTIONS"`
	// MaxConnections is the maximum number of connections.
	MaxConnections int `json:"maxConnections" yaml:"maxConnections" env:"DB_MAX_CONNECTIONS"`
	// MaxLifetime is the maximum time a connection can be open.
	MaxLifetime time.Duration `json:"maxLifetime" yaml:"maxLifetime" env:"DB_MAX_LIFETIME"`
	// BufferPoolSize is the number of query composition buffers to maintain.
	BufferPoolSize int `json:"bufferPoolSize" yaml:"bufferPoolSize" env:"DB_BUFFER_POOL_SIZE"`
}

// WithDSN sets the config dsn and returns a reference to the config.
func (c *Config) WithDSN(dsn string) *Config {
	c.DSN = dsn
	return c
}

// WithHost sets the config host and returns a reference to the config.
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort sets the config host and returns a reference to the config.
func (c *Config) WithPort(port string) *Config {
	c.Port = port
	return c
}

// WithDatabase sets the config database and returns a reference to the config.
func (c *Config) WithDatabase(database string) *Config {
	c.Database = database
	return c
}

// WithSchema sets the config schema and returns a reference to the config.
func (c *Config) WithSchema(schema string) *Config {
	c.Schema = schema
	return c
}

// WithUsername sets the config username and returns a reference to the config.
func (c *Config) WithUsername(username string) *Config {
	c.Username = username
	return c
}

// WithPassword sets the config password and returns a reference to the config.
func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

// WithSSLMode sets the config sslMode and returns a reference to the config.
func (c *Config) WithSSLMode(sslMode string) *Config {
	c.SSLMode = sslMode
	return c
}

// GetDSN returns the postgres dsn (fully quallified url) for the config.
// If unset, it's generated from the host, port and database.
func (c Config) GetDSN(inherited ...string) string {
	return util.Coalesce.String(c.DSN, "", inherited...)
}

// GetHost returns the postgres host for the connection or a default.
func (c Config) GetHost(inherited ...string) string {
	return util.Coalesce.String(c.Host, DefaultHost, inherited...)
}

// GetPort returns the port for a connection if it is not the standard postgres port.
func (c Config) GetPort(inherited ...string) string {
	return util.Coalesce.String(c.Port, DefaultPort, inherited...)
}

// GetDatabase returns the connection database or a default.
func (c Config) GetDatabase(inherited ...string) string {
	return util.Coalesce.String(c.Database, DefaultDatabase, inherited...)
}

// GetSchema returns the connection schema or a default.
func (c Config) GetSchema(inherited ...string) string {
	return util.Coalesce.String(c.Schema, "", inherited...)
}

// GetUsername returns the connection username or a default.
func (c Config) GetUsername(inherited ...string) string {
	return util.Coalesce.String(c.Username, "", inherited...)
}

// GetPassword returns the connection password or a default.
func (c Config) GetPassword(inherited ...string) string {
	return util.Coalesce.String(c.Password, "", inherited...)
}

// GetSSLMode returns the connection ssl mode or a default.
func (c Config) GetSSLMode(inherited ...string) string {
	return util.Coalesce.String(c.SSLMode, DefaultSSLMode, inherited...)
}

// GetUseStatementCache returns if we should enable the statement cache or a default.
func (c Config) GetUseStatementCache(inherited ...bool) bool {
	return util.Coalesce.Bool(c.UseStatementCache, DefaultUseStatementCache, inherited...)
}

// GetIdleConnections returns the number of idle connections or a default.
func (c Config) GetIdleConnections(inherited ...int) int {
	return util.Coalesce.Int(c.IdleConnections, DefaultIdleConnections, inherited...)
}

// GetMaxConnections returns the maximum number of connections or a default.
func (c Config) GetMaxConnections(inherited ...int) int {
	return util.Coalesce.Int(c.MaxConnections, DefaultMaxConnections, inherited...)
}

// GetMaxLifetime returns the maximum lifetime of a driver connection.
func (c Config) GetMaxLifetime(inherited ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.MaxLifetime, DefaultMaxLifetime, inherited...)
}

// GetBufferPoolSize returns the number of query buffers to maintain or a default.
func (c Config) GetBufferPoolSize(inherited ...int) int {
	return util.Coalesce.Int(c.BufferPoolSize, DefaultBufferPoolSize, inherited...)
}

// CreateDSN creates a postgres connection string from the config.
func (c Config) CreateDSN() string {
	if len(c.GetDSN()) != 0 {
		return c.GetDSN()
	}

	var sslMode string
	if len(c.GetSSLMode()) > 0 {
		sslMode = fmt.Sprintf("?sslmode=%s", url.QueryEscape(c.GetSSLMode()))
	}

	var port string
	if len(c.GetPort()) > 0 {
		port = fmt.Sprintf(":%s", c.GetPort())
	}

	if len(c.GetUsername()) > 0 {
		if len(c.GetPassword()) > 0 {
			return fmt.Sprintf("postgres://%s:%s@%s%s/%s%s", url.QueryEscape(c.GetUsername()), url.QueryEscape(c.GetPassword()), c.GetHost(), port, c.GetDatabase(), sslMode)
		}
		return fmt.Sprintf("postgres://%s@%s%s/%s%s", url.QueryEscape(c.GetUsername()), c.GetHost(), port, c.GetDatabase(), sslMode)
	}
	return fmt.Sprintf("postgres://%s%s/%s%s", c.GetHost(), port, c.GetDatabase(), sslMode)
}
