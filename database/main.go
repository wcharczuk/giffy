package main

import (
	"os"
	"strings"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/logger"
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
	conn, err := db.NewFromConfig(&cfg.DB).Open()
	if err != nil {
		log.SyncFatalExit(err)
	}

	var m *migration.Group
	switch strings.ToLower(command) {
	case "migrate":
		m = migrations.Migrations()
	case "init":
		m = initialize.Initialize(&cfg)
	}

	m.WithLogger(log)
	err = m.Apply(conn)
	if err != nil {
		log.SyncFatalExit(err)
	}
}
