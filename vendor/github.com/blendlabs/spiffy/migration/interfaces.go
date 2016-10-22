package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Action is a step in a migration.
type Action func(o *Operation, c *spiffy.DbConnection, tx *sql.Tx) error

// Invocable is a thing that can be invoked.
type Invocable interface {
	Invoke(c *spiffy.DbConnection, tx *sql.Tx) error
}

// Migration is an Invocable that can tested before running.
type Migration interface {
	Label() string
	SetLabel(label string)

	Parent() *Runner
	SetParent(parent *Runner)

	Logger() *Logger
	SetLogger(logger *Logger)

	IsTransactionIsolated() bool

	Test(c *spiffy.DbConnection, optionalTx ...*sql.Tx) error
	Apply(c *spiffy.DbConnection, optionalTx ...*sql.Tx) error
}
