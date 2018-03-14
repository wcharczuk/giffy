package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Invocable is a thing that can be invoked.
type Invocable interface {
	Invoke(c *spiffy.Connection, tx *sql.Tx) error
}

// Action is a function that can be run during a migration step.
type Action func(c *spiffy.Connection, tx *sql.Tx) error
