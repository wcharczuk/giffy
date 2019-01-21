package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// Invocable is a thing that can be invoked.
type Invocable interface {
	Label() string
	Invoke(*Suite, *Group, *db.Connection, *sql.Tx) error
}

// InvocableFunc is a function that can be run during a migration step.
type InvocableFunc func(c *db.Connection, tx *sql.Tx) error

// Statements is an alias to Body(...Statement(stmt))
func Statements(statements ...string) InvocableFunc {
	return func(c *db.Connection, tx *sql.Tx) (err error) {
		for _, statement := range statements {
			err = c.ExecInTx(statement, tx)
			if err != nil {
				return
			}
		}
		return
	}

}

// Actions returns an invocable of a set of actions.
func Actions(actions ...InvocableFunc) InvocableFunc {
	return func(c *db.Connection, tx *sql.Tx) (err error) {
		for _, action := range actions {
			err = action(c, tx)
			if err != nil {
				return err
			}
		}
		return
	}
}

// NoOp performs no action.
func NoOp(c *db.Connection, tx *sql.Tx) error { return nil }
