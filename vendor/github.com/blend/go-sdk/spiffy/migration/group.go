package migration

import (
	"database/sql"
	"fmt"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/spiffy"
)

// New creates a new migration group.
func New(migrations ...Migration) *Group {
	return NewGroup(migrations...)
}

// NewGroup creates a new migration group.
func NewGroup(migrations ...Migration) *Group {
	r := &Group{
		shouldAbortOnError: false,
	}
	r.Add(migrations...)
	return r
}

// Group is an atomic series of migrations.
type Group struct {
	label              string
	shouldAbortOnError bool
	parent             Migration
	stack              []string
	log                *Logger
	migrations         []Migration
}

// Add adds migrations to the suite.
func (g *Group) Add(migrations ...Migration) {
	for _, m := range migrations {
		m.SetParent(g)
		g.migrations = append(g.migrations, m)
	}
}

// Clear removes migrations from the suite.
func (g *Group) Clear() {
	g.migrations = []Migration{}
}

// Label returns a label for the runner.
func (g *Group) Label() string {
	return g.label
}

// SetLabel sets the migration label.
func (g *Group) SetLabel(value string) {
	g.label = value
}

// WithLabel sets the migration label.
func (g *Group) WithLabel(value string) Migration {
	g.label = value
	return g
}

// IsRoot denotes if the runner is the root runner (or not).
func (g *Group) IsRoot() bool {
	return g.parent == nil
}

// Parent returns the runner's parent.
func (g *Group) Parent() Migration {
	return g.parent
}

// SetParent sets the runner's parent.
func (g *Group) SetParent(parent Migration) {
	g.parent = parent
}

// WithParent sets the runner's parent.
func (g *Group) WithParent(parent Migration) Migration {
	g.parent = parent
	return g
}

// ShouldAbortOnError indicates that the group will abort if it sees an error from a step.
func (g *Group) ShouldAbortOnError() bool {
	return g.shouldAbortOnError
}

// SetShouldAbortOnError sets if the group should abort on error.
func (g *Group) SetShouldAbortOnError(value bool) {
	g.shouldAbortOnError = value
}

// WithShouldAbortOnError sets if the group should abort on error.
func (g *Group) WithShouldAbortOnError(value bool) *Group {
	g.shouldAbortOnError = value
	return g
}

// Logger returns the logger.
func (g *Group) Logger() *Logger {
	return g.log
}

// SetLogger sets the logger the Runner should use.
func (g *Group) SetLogger(logger *Logger) {
	g.log = logger
}

// WithLogger sets the logger the Runner should use.
func (g *Group) WithLogger(logger *Logger) Migration {
	g.log = logger
	return g
}

// IsTransactionIsolated returns if the migration is transaction isolated.
func (g *Group) IsTransactionIsolated() bool {
	return true
}

// Test wraps the action in a transaction and rolls the transaction back upon completion.
func (g *Group) Test(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	if g.log != nil {
		g.log.Phase = "test"
	}

	for _, m := range g.migrations {
		if g.log != nil {
			m.SetLogger(g.log)
		}

		err = g.invoke(true, m, c, optionalTx...)
		if err != nil && g.shouldAbortOnError {
			break
		}
	}
	return
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (g *Group) Apply(c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
	if g.log != nil {
		g.log.Phase = "apply"
	}

	for _, m := range g.migrations {
		if g.log != nil {
			m.SetLogger(g.log)
		}

		err = g.invoke(false, m, c, optionalTx...)
		if err != nil && g.shouldAbortOnError {
			break
		}
	}

	if g.IsRoot() && g.log != nil {
		g.log.WriteStats()
	}
	return
}

func (g *Group) invoke(isTest bool, m Migration, c *spiffy.Connection, optionalTx ...*sql.Tx) (err error) {
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
