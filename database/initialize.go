package main

import (
	"log"

	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
	_ "github.com/wcharczuk/giffy/database/initialize"
	_ "github.com/wcharczuk/giffy/database/migrations"
)

func main() {
	err := spiffy.InitDefault(spiffy.NewConnectionFromEnvironment())
	if err != nil {
		log.Fatal(err)
	}

	err = migration.Run(func(suite migration.Migration) error {
		suite.SetLogger(migration.NewLogger())
		return suite.Apply(spiffy.DB())
	})
	if err != nil {
		log.Fatal(err)
	}
}
