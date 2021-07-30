package main

import (
	"os"
	"strings"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/logger"

	"github.com/wcharczuk/giffy/db/initialize"
	"github.com/wcharczuk/giffy/db/migrations"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	command := "migrate"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var cfg config.Giffy
	if err := configutil.Read(&cfg); !configutil.IsIgnored(err) {
		logger.All().SyncFatalExit(err)
	}

	log := logger.NewFromConfig(&cfg.Logger)
	conn := db.MustNewFromConfig(&cfg.DB)
	if err := conn.Open(); err != nil {
		log.SyncFatalExit(err)
	}

	var m *migration.Suite
	switch strings.ToLower(command) {
	case "migrate":
		m = migrations.Migrations()
	case "init":
		m = initialize.Initialize(&cfg)
	}

	m.WithLogger(log)
	if err := m.Apply(conn); err != nil {
		log.SyncFatalExit(err)
	}
}
