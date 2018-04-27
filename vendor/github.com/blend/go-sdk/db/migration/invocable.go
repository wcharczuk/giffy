package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// Statements is an alias to Body(...Statement(stmt))
func Statements(stmts ...string) Invocable {
	return statements(stmts)
}

// statements is a collection of statements to run as an action.
// they are executed serially.
type statements []string

// Invoke executes the statement block
func (s statements) Invoke(c *db.Connection, tx *sql.Tx) (err error) {
	for _, step := range s {
		err = c.ExecInTx(step, tx)
		if err != nil {
			return
		}
	}
	return
}

// Actions returns an invocable of a set of actions.
func Actions(invocableActions ...InvocableAction) Invocable {
	return actions(invocableActions)
}

// actions wraps a user supplied invocation body.
type actions []InvocableAction

// Invoke applies the invocation.
func (a actions) Invoke(c *db.Connection, tx *sql.Tx) error {
	var err error
	for _, action := range a {
		err = action(c, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	// NoOp is used when testing guards.
	NoOp = noOp{}
)

type noOp struct{}

// Invoke implements Invocable.
func (no noOp) Invoke(c *db.Connection, tx *sql.Tx) error { return nil }
