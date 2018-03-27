package main

import (
	"fmt"
	"os"
	"strings"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/configutil"
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
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
	err := configutil.Read(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	log := logger.NewFromConfig(&cfg.Logger)
	db, err := spiffy.NewFromConfig(&cfg.DB).Open()
	if err != nil {
		log.Fatal(err)
	}

	var migration migration.Migration

	switch strings.ToLower(command) {
	case "migrate":
		migration = migrations.Migrations()
	case "init":
		migration = initialize.Initialize(&cfg)
	}

	err = migration.Apply(db)
	if err != nil {
		log.Fatal(err)
	}
}
