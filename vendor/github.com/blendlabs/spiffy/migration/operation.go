package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Step is an alias to NewOperation.
func Step(action Action, body Statement, args ...string) *Operation {
	return NewOperation(action, body, args...)
}

// NewOperation creates a new invocable.
func NewOperation(action Action, body Statement, args ...string) *Operation {
	return &Operation{
		action: action,
		body:   body,
		args:   args,
	}
}

// Operation is a closure for a Operation
type Operation struct {
	label  string
	parent *Runner
	logger *Logger
	action Action
	body   Statement
	args   []string
}

// Label returns the operation label.
func (o *Operation) Label() string {
	return o.label
}

// SetLabel sets the operation label.
func (o *Operation) SetLabel(label string) {
	o.label = label
}

// Parent returns the parent.
func (o *Operation) Parent() *Runner {
	return o.parent
}

// SetParent sets the operation parent.
func (o *Operation) SetParent(parent *Runner) {
	o.parent = parent
}

// Logger returns the logger
func (o *Operation) Logger() *Logger {
	return o.logger
}

// SetLogger implements the migration method `SetLogger`.
func (o *Operation) SetLogger(logger *Logger) {
	o.logger = logger
}

// IsTransactionIsolated returns if this migration requires it's own migration
func (o *Operation) IsTransactionIsolated() bool {
	return false
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (o *Operation) Test(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	err = o.Apply(c, optionalTx...)
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (o *Operation) Apply(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	tx := spiffy.OptionalTx(optionalTx...)
	err = o.action(o, c, tx)
	return
}
