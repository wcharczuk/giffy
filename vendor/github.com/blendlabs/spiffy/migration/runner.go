package migration

import (
	"database/sql"
	"log"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
)

// New creates a new migration series.
func New(name string, migrations ...Migration) Migration {
	return &Runner{
		Name:       name,
		Migrations: migrations,
	}
}

// Runner runs the migrations
type Runner struct {
	Name       string
	Logger     *log.Logger
	Migrations []Migration
}

// SetLogger sets the logger the Runner should use.
func (r *Runner) SetLogger(logger *log.Logger) {
	r.Logger = logger
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

	setLoggerPhase(r.Logger, "test", r.Name)
	defer setLoggerPhase(r.Logger, util.StringEmpty, util.StringEmpty)

	err = r.Invoke(c, tx)
	if err != nil {
		logError(r.Logger, err)
	}
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

	setLoggerPhase(r.Logger, "apply", r.Name)
	defer setLoggerPhase(r.Logger, util.StringEmpty, util.StringEmpty)

	err = r.Invoke(c, tx)
	if err != nil {
		logError(r.Logger, err)
	}
	return
}

// Invoke runs the suite against a given connection and transaction.
func (r Runner) Invoke(c *spiffy.DbConnection, tx *sql.Tx) (err error) {
	for _, m := range r.Migrations {
		if r.Logger != nil {
			m.SetLogger(r.Logger)
		}
		err = m.Invoke(c, tx)
		if err != nil {
			return err
		}
	}
	return nil
}
