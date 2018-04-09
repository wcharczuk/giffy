package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// NewStep is an alias to NewOperation.
func NewStep(guard Guard, body Invocable) *Step {
	return &Step{
		guard: guard,
		body:  body,
	}
}

// Step is a single guarded function.
type Step struct {
	label  string
	parent Migration
	logger *Logger

	guard Guard
	body  Invocable
}

// Label returns the operation label.
func (s *Step) Label() string {
	return s.label
}

// SetLabel sets the operation label.
func (s *Step) SetLabel(label string) {
	s.label = label
}

// WithLabel sets the operation label.
func (s *Step) WithLabel(label string) Migration {
	s.label = label
	return s
}

// Parent returns the parent.
func (s *Step) Parent() Migration {
	return s.parent
}

// SetParent sets the operation parent.
func (s *Step) SetParent(parent Migration) {
	s.parent = parent
}

// WithParent sets the operation parent.
func (s *Step) WithParent(parent Migration) Migration {
	s.parent = parent
	return s
}

// Logger returns the logger
func (s *Step) Logger() *Logger {
	return s.logger
}

// SetLogger implements the migration method `SetLogger`.
func (s *Step) SetLogger(logger *Logger) {
	s.logger = logger
}

// WithLogger implements the migration method `WithLogger`.
func (s *Step) WithLogger(logger *Logger) Migration {
	s.logger = logger
	return s
}

// IsTransactionIsolated returns if this migration requires its own transaction.
func (s *Step) IsTransactionIsolated() bool {
	return false
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (s *Step) Test(c *db.Connection, optionalTx ...*sql.Tx) (err error) {
	err = s.Apply(c, optionalTx...)
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (s *Step) Apply(c *db.Connection, txs ...*sql.Tx) (err error) {
	tx := db.OptionalTx(txs...)
	err = s.guard(s, c, tx)
	return
}
