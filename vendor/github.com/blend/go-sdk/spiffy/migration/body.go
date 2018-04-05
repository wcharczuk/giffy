package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/spiffy"
)

// Body returns an invocable of a set of invocable actions.
func Body(actions ...Action) Invocable {
	return &body{actions: actions}
}

// body wraps a user supplied invocation body.
type body struct {
	actions []Action
}

// Invoke applies the invocation.
func (b *body) Invoke(c *spiffy.Connection, tx *sql.Tx) error {
	var err error
	for _, action := range b.actions {
		err = action(c, tx)
		if err != nil {
			return err
		}
	}
	return nil
}
