package migration

import (
	"database/sql"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

// Step is an alias to NewOperation.
func Step(action Action, body Statement, args ...string) Migration {
	return NewOperation(action, body, args...)
}

// NewOperation creates a new invocable.
func NewOperation(action Action, body Statement, args ...string) Migration {
	return &Operation{
		Action: action,
		Body:   body,
		Args:   args,
	}
}

// Operation is a closure for a Operation
type Operation struct {
	Stack  []string
	Logger *Logger
	Action Action
	Body   Statement
	Args   []string
}

// Logged implements the migration method `Logged`.
func (o *Operation) Logged(logger *Logger, stack ...string) {
	o.Stack = append([]string{}, stack...)
	o.Logger = logger
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (o Operation) Test(c *spiffy.DbConnection) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return
	}
	defer func() {
		err = exception.Wrap(tx.Rollback())
	}()
	err = o.Invoke(c, tx)
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (o Operation) Apply(c *spiffy.DbConnection) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = exception.Wrap(tx.Commit())
		} else {
			err = exception.WrapMany(err, exception.New(tx.Rollback()))
		}
	}()

	err = o.Invoke(c, tx)
	return
}

// Invoke runs the operation against the given connection and transaction.
func (o Operation) Invoke(c *spiffy.DbConnection, tx *sql.Tx) error {
	return o.Action(o.Stack, o.Logger, c, tx, o.Body, o.Args...)
}
