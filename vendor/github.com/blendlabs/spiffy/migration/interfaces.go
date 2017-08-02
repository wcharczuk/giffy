package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Migration is either a group of steps or the entire suite.
type Migration interface {
	Label() string
	SetLabel(label string)

	Parent() Migration
	SetParent(parent Migration)

	Logger() *Logger
	SetLogger(logger *Logger)

	IsTransactionIsolated() bool

	Test(c *spiffy.Connection, optionalTx ...*sql.Tx) error
	Apply(c *spiffy.Connection, optionalTx ...*sql.Tx) error
}

// GuardAction is a control for migration steps.
type GuardAction func(o *Operation, c *spiffy.Connection, tx *sql.Tx) error

// Invocable is a thing that can be invoked.
type Invocable interface {
	Invoke(c *spiffy.Connection, tx *sql.Tx) error
}

// InvocableAction is a function that can be run during a migration step.
type InvocableAction func(c *spiffy.Connection, tx *sql.Tx) error
