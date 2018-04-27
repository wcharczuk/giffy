package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

var (
	// assert step implements migration.
	_ Migration = &Step{}
)

// NewStep is an alias to NewOperation.
func NewStep(guard Guard, body Invocable) *Step {
	return &Step{
		transactionBound: true,
		guard:            guard,
		body:             body,
	}
}

// Step is a single guarded function.
type Step struct {
	label            string
	parent           Migration
	collector        *Collector
	transactionBound bool
	guard            Guard
	body             Invocable
}

// WithLabel sets the operation label.
func (s *Step) WithLabel(label string) Migration {
	s.label = label
	return s
}

// Label returns the operation label.
func (s *Step) Label() string {
	return s.label
}

// WithParent sets the operation parent.
func (s *Step) WithParent(parent Migration) Migration {
	s.parent = parent
	return s
}

// Parent returns the parent.
func (s *Step) Parent() Migration {
	return s.parent
}

// WithCollector sets the collector.
func (s *Step) WithCollector(collector *Collector) Migration {
	s.collector = collector
	return s
}

// Collector returns the collector
func (s *Step) Collector() *Collector {
	return s.collector
}

// WithTransactionBound sets if the migration manages its own transactions or not.
func (s *Step) WithTransactionBound(transactionBound bool) Migration {
	s.transactionBound = transactionBound
	return s
}

// TransactionBound returns if the migration manages its own transactions.
func (s *Step) TransactionBound() bool {
	return s.transactionBound
}

// Apply wraps the action in a transaction and commits it if there were no errors, rolling back if there were.
func (s *Step) Apply(c *db.Connection, txs ...*sql.Tx) (err error) {
	err = s.guard(s, c, db.OptionalTx(txs...))
	return
}
