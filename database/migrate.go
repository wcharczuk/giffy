package main

package main

import (
	"log"

	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
	_ "github.com/wcharczuk/giffy/db/migrations"
)

func main() {
	err := spiffy.SetDefaultDb(spiffy.NewDbConnectionFromEnvironment())
	if err != nil {
		log.Fatal(err)
	}

	err = migration.Run(func(suite migration.Migration) error {
		suite.SetLogger(migration.NewLogger())
		return suite.Apply(spiffy.DefaultDb())
	})
	if err != nil {
		log.Fatal(err)
	}
}
