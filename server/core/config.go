package core

import (
	"fmt"
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
