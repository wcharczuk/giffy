package core

import (
	"fmt"
	"net"
	"os"

	"github.com/blendlabs/spiffy"
)

const (
	RequestLogFormat = "datetime c-ip cs-method cs-uri cs-status time-taken bytes"
)

type DBConfig struct {
	Server   string
	Schema   string
	User     string
	Password string
}

func (db *DBConfig) InitFromEnvironment() {
	db.Server = os.Getenv("DB_HOST")
	db.Schema = os.Getenv("DB_SCHEMA")
	db.User = os.Getenv("DB_USER")
	db.Password = os.Getenv("DB_PASSWORD")
}

func (db DBConfig) AsPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", db.User, db.Password, db.Server, db.Schema)
}

func SetupDatabaseContext(config *DBConfig) error {
	spiffy.CreateDbAlias("main", spiffy.NewDbConnection(config.Server, config.Schema, config.User, config.Password))
	spiffy.SetDefaultAlias("main")

	_, dbError := spiffy.DefaultDb().Open()
	if dbError != nil {
		return dbError
	}

	spiffy.DefaultDb().Connection.SetMaxIdleConns(50)
	return nil
}

func DBInit() error {
	config := &DBConfig{}
	config.InitFromEnvironment()
	return SetupDatabaseContext(config)
}

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
