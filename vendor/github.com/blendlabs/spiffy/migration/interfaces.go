package migration

import (
	"database/sql"
	"log"

	"github.com/blendlabs/spiffy"
)

// Action is a step in a migration.
type Action func(l *log.Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error

// Invocable is a thing that can be invoked.
type Invocable interface {
	Invoke(c *spiffy.DbConnection, tx *sql.Tx) error
}

// Migration is an Invocable that can tested before running.
type Migration interface {
	SetLogger(logger *log.Logger)
	Test(c *spiffy.DbConnection) error
	Apply(c *spiffy.DbConnection) error
	Invoke(c *spiffy.DbConnection, tx *sql.Tx) error
}
