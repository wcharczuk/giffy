package migration

import (
	"database/sql"
	"fmt"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

// New creates a new migration series.
func New(label string, migrations ...Migration) *Suite {
	r := &Suite{
		label: label,
	}
	r.Add(migrations...)
	return r
}

// Suite runs the migrations
type Suite struct {
	label              string
	parent             Migration
	shouldAbortOnError bool
	stack              []string
	logger             *Logger
	migrations         []Migration
}

// Add adds migrations to the suite.
func (s *Suite) Add(migrations ...Migration) {
	for _, m := range migrations {
		m.SetParent(s)
		s.migrations = append(s.migrations, m)
	}
}

// Clear removes migrations from the suite.
func (s *Suite) Clear() {
	s.migrations = []Migration{}
}

// Label returns a label for the runner.
func (s *Suite) Label() string {
	return s.label
}

// SetLabel sets the migration label.
func (s *Suite) SetLabel(value string) {
	s.label = value
}

// IsRoot denotes if the runner is the root runner (or not).
func (s *Suite) IsRoot() bool {
	return s.parent == nil
}

// Parent returns the runner's parent.
func (s *Suite) Parent() Migration {
	return s.parent
}

// SetParent sets the runner's parent.
func (s *Suite) SetParent(parent Migration) {
	s.parent = parent
}

// ShouldAbortOnError indicates that the runner will abort if it sees an error from a step.
func (s *Suite) ShouldAbortOnError() bool {
	return s.shouldAbortOnError
}

// SetShouldAbortOnError sets if the runner should abort on error.
func (s *Suite) SetShouldAbortOnError(value bool) {
	s.shouldAbortOnError = value
}

// Logger returns the logger.
func (s *Suite) Logger() *Logger {
	return s.logger
}

// SetLogger sets the logger the Runner should use.
func (s *Suite) SetLogger(logger *Logger) {
	s.logger = logger
}

// IsTransactionIsolated returns if the migration is transaction isolated.
func (s *Suite) IsTransactionIsolated() bool {
	return true
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (s *Suite) Test(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	if s.logger != nil {
		s.logger.Phase = "test"
	}

	for _, m := range s.migrations {
		if s.logger != nil {
			m.SetLogger(s.logger)
		}

		err = s.invokeMigration(true, m, c, optionalTx...)
		if err != nil && s.shouldAbortOnError {
			break
		}
	}
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (s *Suite) Apply(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	if s.logger != nil {
		s.logger.Phase = "apply"
	}

	for _, m := range s.migrations {
		if s.logger != nil {
			m.SetLogger(s.logger)
		}

		err = s.invokeMigration(false, m, c, optionalTx...)
		if err != nil && s.shouldAbortOnError {
			break
		}
	}

	if s.IsRoot() && s.logger != nil {
		s.logger.WriteStats()
	}
	return
}

func (s *Suite) invokeMigration(isTest bool, m Migration, c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
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
			err = exception.Nest(err, exception.New(tx.Rollback()))
		}
	}()
	err = m.Apply(c, tx)
	return
}
