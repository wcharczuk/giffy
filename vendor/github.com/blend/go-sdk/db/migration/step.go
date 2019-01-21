package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// NewStep is an alias to NewOperation.
func NewStep(guard GuardFunc, body InvocableFunc) *Step {
	return &Step{
		guard: guard,
		body:  body,
	}
}

// Step is a single guarded function.
type Step struct {
	label string
	guard GuardFunc
	body  InvocableFunc
}

// WithLabel sets the operation label.
func (s *Step) WithLabel(label string) *Step {
	s.label = label
	return s
}

// Label returns the operation label.
func (s *Step) Label() string {
	return s.label
}

// Apply applies a step in isolation.
func (s *Step) Apply(c *db.Connection, tx *sql.Tx) error {
	return s.Invoke(nil, nil, c, tx)
}

// Invoke runs the body if the provided guard passes.
func (s *Step) Invoke(suite *Suite, group *Group, c *db.Connection, tx *sql.Tx) (err error) {
	err = s.guard(suite, group, s, c, tx)
	return
}
