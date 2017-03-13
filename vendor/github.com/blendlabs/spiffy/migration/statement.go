package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Body is an alias to NewStatement.
func Body(statements ...string) Statement {
	return NewStatement(statements...)
}

// NewStatement creates a statement block.
func NewStatement(statements ...string) Statement {
	return Statement(statements)
}

// Statement is an atomic unit of work. It can be multiple individual sql statements.
// This is what is run by the operation gates (if index exists / if column exists etc.)
type Statement []string

// Invoke executes the statement block
func (s Statement) Invoke(c *spiffy.Connection, tx *sql.Tx) (err error) {
	for _, step := range s {
		err = c.ExecInTx(step, tx)
		if err != nil {
			return
		}
	}
	return
}
