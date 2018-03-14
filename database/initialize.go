package main

import (
	"log"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
	_ "github.com/wcharczuk/giffy/database/initialize"
	_ "github.com/wcharczuk/giffy/database/migrations"
)

func main() {
	err := spiffy.OpenDefault(spiffy.NewFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	migration.Default().WithLogger(migration.NewLogger(logger.NewFromEnv()))
	err = migration.Default().Apply(spiffy.Default())
	if err != nil {
		log.Fatal(err)
	}
}
