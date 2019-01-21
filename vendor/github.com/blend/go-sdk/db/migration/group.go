package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
)

// NewGroup creates a new migration group.
func NewGroup(steps ...Invocable) *Group {
	r := &Group{}
	return r.WithSteps(steps...)
}

// Group is an atomic series of migrations.
// It uses transactions to apply a set of sub-migrations as a unit.
type Group struct {
	label string
	steps []Invocable
}

// WithSteps adds steps to the group.
func (g *Group) WithSteps(steps ...Invocable) *Group {
	for _, s := range steps {
		g.steps = append(g.steps, s)
	}
	return g
}

// WithLabel sets the migration label.
func (g *Group) WithLabel(value string) *Group {
	g.label = value
	return g
}

// Label returns a label for the runner.
func (g *Group) Label() string {
	return g.label
}

// Invoke runs the steps in a transaction.
func (g *Group) Invoke(suite *Suite, c *db.Connection) (err error) {
	var tx *sql.Tx
	tx, err = c.Begin()
	if err != nil {
		return
	}

	// commit or rollback the transaction.
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = exception.Nest(err, txErr)
			}
		} else {
			if txErr := tx.Commit(); txErr != nil {
				err = exception.Nest(err, txErr)
			}
		}
	}()

	for _, s := range g.steps {
		err = s.Invoke(suite, g, c, tx)
		if err != nil {
			return
		}
	}

	return
}
