package main

import (
	"os"
	"strings"

	"github.com/blend/go-sdk/configutil"
	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/spiffy"
	"github.com/blend/go-sdk/spiffy/migration"
	"github.com/wcharczuk/giffy/database/initialize"
	"github.com/wcharczuk/giffy/database/migrations"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	command := "migrate"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var cfg config.Giffy
	if err := configutil.Read(&cfg); err != nil {
		logger.All().SyncFatalExit(err)
	}

	log := logger.NewFromConfig(&cfg.Logger)
	db, err := spiffy.NewFromConfig(&cfg.DB).Open()
	if err != nil {
		log.SyncFatalExit(err)
	}

	var m migration.Migration
	switch strings.ToLower(command) {
	case "migrate":
		m = migrations.Migrations()
	case "init":
		m = initialize.Initialize(&cfg)
	}

	m.WithLogger(migration.NewLogger(log))
	err = m.Apply(db)
	if err != nil {
		log.SyncFatalExit(err)
	}
}
