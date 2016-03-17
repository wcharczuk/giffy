package core

import (
	"fmt"
	"net"
	"os"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
)

const (
	// RequestLogFormat is the default log format.
	RequestLogFormat = "datetime c-ip cs-method cs-uri cs-status time-taken bytes"
)

// DBConfig is the basic config object for db connections.
type DBConfig struct {
	Server   string
	Schema   string
	User     string
	Password string

	dsn string
}

// InitFromEnvironment initializes the db config from environment variables.
func (db *DBConfig) InitFromEnvironment() {
	dsn := os.Getenv("DATABASE_URL")
	if len(dsn) != 0 {
		db.InitFromDSN(dsn)
	} else {
		db.Server = os.Getenv("DB_HOST")
		db.Schema = os.Getenv("DB_SCHEMA")
		db.User = os.Getenv("DB_USER")
		db.Password = os.Getenv("DB_PASSWORD")
	}
}

// InitFromDSN initializes the db config from a dsn.
func (db *DBConfig) InitFromDSN(dsn string) {
	db.dsn = dsn
}

// DSN returns the config as a postgres dsn.
func (db DBConfig) DSN() string {
	if len(db.dsn) != 0 {
		return db.dsn
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", db.User, db.Password, db.Server, db.Schema)
}

// SetupDatabaseContext writes the config to spiffy.
func SetupDatabaseContext(config *DBConfig) error {
	spiffy.CreateDbAlias("main", spiffy.NewDbConnectionFromDSN(config.DSN()))
	spiffy.SetDefaultAlias("main")

	_, dbError := spiffy.DefaultDb().Open()
	if dbError != nil {
		return dbError
	}

	spiffy.DefaultDb().Connection.SetMaxIdleConns(50)
	return nil
}

// DBInit reads the config from the environment and sets up spiffy.
func DBInit() error {
	config := &DBConfig{}
	config.InitFromEnvironment()
	return SetupDatabaseContext(config)
}

// ConfigPort is the port the server should listen on.
func ConfigPort() string {
	envPort := os.Getenv("PORT")
	if !util.IsEmpty(envPort) {
		return envPort
	}
	return "8080"
}

// ConfigLocalIP is the server local IP.
func ConfigLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var configKey []byte

// ConfigKey is the app secret we use to encrypt things.
func ConfigKey() []byte {
	if configKey == nil {
		keyBlob := os.Getenv("ENCRYPTION_KEY")
		if len(keyBlob) != 0 {
			key, keyErr := util.Base64Decode(keyBlob)
			if keyErr != nil {
				println(keyErr.Error())
				return key
			}
			configKey = key
		}
	}
	return configKey
}

// ConfigEnvironment returns the current environment.
func ConfigEnvironment() string {
	env := os.Getenv("ENV")
	if len(env) != 0 {
		return env
	}
	return "dev"
}

// ConfigHostname returns the hostname for the server.
func ConfigHostname() string {
	envHost := os.Getenv("HOSTNAME")
	if len(envHost) != 0 {
		return envHost
	}

	if ConfigEnvironment() == "dev" {
		return "dev.giffy.charczuk.com"
	}
	return "giffy.charczuk.com"
}

// ConfigGoogleClientID returns the google client id.
func ConfigGoogleClientID() string {
	return os.Getenv("GOOGLE_CLIENT_ID")
}

// ConfigGoogleSecret returns the google secret.
func ConfigGoogleSecret() string {
	return os.Getenv("GOOGLE_CLIENT_SECRET")
}
