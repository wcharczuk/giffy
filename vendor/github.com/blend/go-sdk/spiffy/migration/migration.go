package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/spiffy"
)

// Migration is either a group of steps or the entire suite.
type Migration interface {
	Label() string
	SetLabel(label string)
	WithLabel(label string) Migration

	Parent() Migration
	SetParent(parent Migration)
	WithParent(parent Migration) Migration

	Logger() *Logger
	SetLogger(logger *Logger)
	WithLogger(logger *Logger) Migration

	IsTransactionIsolated() bool

	Test(c *spiffy.Connection, optionalTx ...*sql.Tx) error
	Apply(c *spiffy.Connection, optionalTx ...*sql.Tx) error
}
