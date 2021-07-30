package main

import (
	"context"
	"os"
	"strings"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/logger"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

func main() {
	command := "migrate"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var cfg config.Giffy
	if _, err := configutil.Read(&cfg); !configutil.IsIgnored(err) {
		logger.FatalExit(err)
	}

	log := logger.MustNew(logger.OptConfig(cfg.Logger))
	conn := db.MustNew(
		db.OptConfig(cfg.DB),
		db.OptLog(log),
	)
	if err := conn.Open(); err != nil {
		logger.FatalExit(err)
	}

	var m *migration.Suite
	switch strings.ToLower(command) {
	case "migrate":
		m = model.Migrations(&cfg)
	case "init":
		m = model.Schema(&cfg)
	}
	m.Log = log

	if err := m.Apply(context.Background(), conn); err != nil {
		logger.MaybeFatalExit(log, err)
	}
}
