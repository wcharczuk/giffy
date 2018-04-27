package migration

import (
	"database/sql"
	"fmt"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

var (
	// Assert group implements migration.
	_ Migration = &Group{}
)

// NewGroup creates a new migration group.
func NewGroup(migrations ...Migration) *Group {
	r := &Group{
		useTransaction: true,
		abortOnError:   true,
	}
	return r.With(migrations...)
}

// Group is an atomic series of migrations.
// It uses transactions to apply a set of sub-migrations as a unit.
type Group struct {
	label              string
	abortOnError       bool
	rollbackOnComplete bool
	useTransaction     bool
	parent             Migration
	collector          *Collector
	migrations         []Migration
}

// With adds migrations to the group and returns a reference.
func (g *Group) With(migrations ...Migration) *Group {
	g.Add(migrations...)
	return g
}

// Add adds migrations to the group.
func (g *Group) Add(migrations ...Migration) {
	for _, m := range migrations {
		g.migrations = append(g.migrations, m.WithParent(g))
	}
}

// WithLabel sets the migration label.
func (g *Group) WithLabel(value string) Migration {
	g.label = value
	return g
}

// Label returns a label for the runner.
func (g *Group) Label() string {
	return g.label
}

// IsRoot denotes if the runner is the root runner (or not).
func (g *Group) IsRoot() bool {
	return g.parent == nil
}

// WithParent sets the runner's parent.
func (g *Group) WithParent(parent Migration) Migration {
	g.parent = parent
	return g
}

// Parent returns the runner's parent.
func (g *Group) Parent() Migration {
	return g.parent
}

// WithAbortOnError sets if the group should abort on error.
func (g *Group) WithAbortOnError(value bool) *Group {
	g.abortOnError = value
	return g
}

// AbortOnError indicates that the group will abort if it sees an error from a step.
func (g *Group) AbortOnError() bool {
	return g.abortOnError
}

// WithCollector sets the collector.
func (g *Group) WithCollector(collector *Collector) Migration {
	g.collector = collector
	return g
}

// Collector returns the collector.
func (g *Group) Collector() *Collector {
	return g.collector
}

// WithLogger sets the collector output logger.
func (g *Group) WithLogger(log *logger.Logger) *Group {
	if g.collector == nil {
		g.collector = &Collector{
			output: log,
		}
		return g
	}
	g.collector.output = log
	return g
}

// WithUseTransaction sets if we should begin a transaction for the work within the group.
func (g *Group) WithUseTransaction(useTransaction bool) Migration {
	g.useTransaction = useTransaction
	return g
}

// UseTransaction returns if the group should wrap child steps in a transaction.
func (g *Group) UseTransaction() bool {
	return g.useTransaction
}

// TransactionBound returns if the migration manages its own transactions.
// This is a group, it is the one doing the managing.
func (g *Group) TransactionBound() bool {
	return false
}

// WithRollbackOnComplete sets if we should roll the transaction back on complete.
func (g *Group) WithRollbackOnComplete(rollbackOnComplete bool) *Group {
	g.rollbackOnComplete = rollbackOnComplete
	return g
}

// RollbackOnComplete returns if we should rollback the migration on complete.
func (g *Group) RollbackOnComplete() bool {
	return g.rollbackOnComplete
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (g *Group) Apply(c *db.Connection, txs ...*sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", err)
		}

		if g.IsRoot() && g.collector != nil {
			g.collector.WriteStats()
		}
	}()

	// if the migration
	var tx *sql.Tx
	if g.useTransaction {
		tx, err = c.Begin()
		if err != nil {
			return
		}

		defer func() {
			if err != nil || g.rollbackOnComplete {
				err = exception.Nest(err, exception.New(tx.Rollback()))
			} else if err == nil {
				err = exception.Wrap(tx.Commit())
			}
		}()
	}

	for _, m := range g.migrations {
		// if the migration is a group or something else that manages it's own transactions ...
		if m.TransactionBound() {
			err = m.WithCollector(g.collector).Apply(c, tx)
		} else {
			err = m.WithCollector(g.collector).Apply(c)
		}
		if err != nil && g.abortOnError {
			return
		}
		continue
	}

	return
}
