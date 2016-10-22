package migration

import (
	"database/sql"
	"fmt"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

// New creates a new migration series.
func New(label string, migrations ...Migration) *Runner {
	r := &Runner{
		label: label,
	}
	r.addMigrations(migrations...)
	return r
}

// Runner runs the migrations
type Runner struct {
	label              string
	parent             *Runner
	shouldAbortOnError bool
	stack              []string
	logger             *Logger
	migrations         []Migration
}

func (r *Runner) addMigrations(migrations ...Migration) {
	for _, m := range migrations {
		m.SetParent(r)
		r.migrations = append(r.migrations, m)
	}
}

// Label returns a label for the runner.
func (r *Runner) Label() string {
	return r.label
}

// SetLabel sets the migration label.
func (r *Runner) SetLabel(value string) {
	r.label = value
}

// IsRoot denotes if the runner is the root runner (or not).
func (r *Runner) IsRoot() bool {
	return r.parent == nil
}

// Parent returns the runner's parent.
func (r *Runner) Parent() *Runner {
	return r.parent
}

// SetParent sets the runner's parent.
func (r *Runner) SetParent(parent *Runner) {
	r.parent = parent
}

// ShouldAbortOnError indicates that the runner will abort if it sees an error from a step.
func (r *Runner) ShouldAbortOnError() bool {
	return r.shouldAbortOnError
}

// SetShouldAbortOnError sets if the runner should abort on error.
func (r *Runner) SetShouldAbortOnError(value bool) {
	r.shouldAbortOnError = value
}

// Logger returns the logger.
func (r *Runner) Logger() *Logger {
	return r.logger
}

// SetLogger sets the logger the Runner should use.
func (r *Runner) SetLogger(logger *Logger) {
	r.logger = logger
}

// IsTransactionIsolated returns if the migration is transaction isolated.
func (r *Runner) IsTransactionIsolated() bool {
	return true
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (r *Runner) Test(c *spiffy.DbConnection, optionalTx ...*sql.Tx) (err error) {
	if r.logger != nil {
		r.logger.Phase = "test"
	}

	for _, m := range r.migrations {
		if r.logger != nil {
			m.SetLogger(r.logger)
		}

		err = r.invokeMigration(true, m, c, optionalTx...)
		if err != nil && r.shouldAbortOnError {
			break
		}
	}
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (r *Runner) Apply(c *spiffy.DbConnection, optionalTx ...*sql.Tx) (err error) {
	if r.logger != nil {
		r.logger.Phase = "apply"
	}

	for _, m := range r.migrations {
		if r.logger != nil {
			m.SetLogger(r.logger)
		}

		err = r.invokeMigration(false, m, c, optionalTx...)
		if err != nil && r.shouldAbortOnError {
			break
		}
	}

	if r.IsRoot() && r.logger != nil {
		r.logger.WriteStats()
	}
	return
}

func (r *Runner) invokeMigration(isTest bool, m Migration, c *spiffy.DbConnection, optionalTx ...*sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", err)
		}
	}()

	if m.IsTransactionIsolated() {
		err = m.Apply(c, spiffy.OptionalTx(optionalTx...))
		return
	}

	var tx *sql.Tx
	tx, err = c.Begin()
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
	err = m.Apply(c, tx)
	return
}
