package migration

import (
	"database/sql"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

// New creates a new migration series.
func New(name string, migrations ...Migration) Migration {
	return &Runner{
		Name:       name,
		Stack:      []string{name},
		Migrations: migrations,
	}
}

// Runner runs the migrations
type Runner struct {
	Name               string
	ShouldBreakOnError bool
	Stack              []string
	Logger             *Logger
	Migrations         []Migration
	IsDefault          bool
}

// Logged sets the logger the Runner should use.
func (r *Runner) Logged(logger *Logger, stack ...string) {
	r.Logger = logger
	r.Stack = append(stack, r.Stack...)
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (r Runner) Test(c *spiffy.DbConnection) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = exception.Wrap(tx.Rollback())
	}()
	r.Logger.Phase = "test"
	err = r.Invoke(c, tx)
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (r Runner) Apply(c *spiffy.DbConnection) (err error) {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = exception.Wrap(tx.Commit())
		} else {
			err = exception.WrapMany(err, exception.New(tx.Rollback()))
		}
	}()
	if r.Logger != nil {
		r.Logger.Phase = "apply"
	}
	err = r.Invoke(c, tx)
	return
}

// Invoke runs the suite against a given connection and transaction.
func (r Runner) Invoke(c *spiffy.DbConnection, tx *sql.Tx) (err error) {
	for _, m := range r.Migrations {
		if r.Logger != nil {
			m.Logged(r.Logger, r.Stack...)
		}
		err = m.Invoke(c, tx)
		if err != nil && r.ShouldBreakOnError {
			break
		}
	}

	if r.IsDefault && r.Logger != nil {
		r.Logger.WriteStats()
	}
	return
}
