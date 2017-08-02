package main

import (
	"log"

	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
	_ "github.com/wcharczuk/giffy/database/initialize"
	_ "github.com/wcharczuk/giffy/database/migrations"
)

func main() {
	err := spiffy.OpenDefault(spiffy.NewConnectionFromEnvironment())
	if err != nil {
		log.Fatal(err)
	}

	migration.Default().SetLogger(migration.NewLogger())
	err = migration.Default().Apply(spiffy.Default())
	if err != nil {
		log.Fatal(err)
	}
}
