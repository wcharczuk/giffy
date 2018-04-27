package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// Guard is a control for migration steps.
type Guard func(s *Step, c *db.Connection, tx *sql.Tx) error

// Invocable is a thing that can be invoked.
type Invocable interface {
	Invoke(c *db.Connection, tx *sql.Tx) error
}

// InvocableAction is a function that can be run during a migration step.
type InvocableAction func(c *db.Connection, tx *sql.Tx) error

// Migration is either a group of steps or the entire suite.
type Migration interface {
	WithLabel(label string) Migration
	Label() string

	WithParent(parent Migration) Migration
	Parent() Migration

	WithCollector(*Collector) Migration
	Collector() *Collector

	// TransactionBound indicates if this migration manages its transaction context.
	// If this returns true, a parent will not pass a transaction into Apply.
	// If it is false, a transaction will be started for this
	TransactionBound() bool

	// Apply runs the migration.
	Apply(c *db.Connection, txs ...*sql.Tx) error
}
