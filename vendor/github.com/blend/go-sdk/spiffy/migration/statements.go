package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/spiffy"
)

// Statements is an alias to Body(...Statement(stmt))
func Statements(stmts ...string) Invocable {
	return statements(stmts)
}

// statements is a collection of statements to run as an action.
// they are executed serially.
type statements []string

// Invoke executes the statement block
func (s statements) Invoke(c *spiffy.Connection, tx *sql.Tx) (err error) {
	for _, step := range s {
		err = c.ExecInTx(step, tx)
		if err != nil {
			return
		}
	}
	return
}
