package migration

import (
	"fmt"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

const (
	// StatApplied is a stat name.
	StatApplied = "applied"
	// StatFailed is a stat name.
	StatFailed = "failed"
	// StatSkipped is a stat name.
	StatSkipped = "skipped"
	// StatTotal is a stat name.
	StatTotal = "total"
)

// New returns a new suite of groups.
func New(groups ...*Group) *Suite {
	return &Suite{
		groups: groups,
	}
}

// Suite is a migration suite.
type Suite struct {
	log    logger.Log
	groups []*Group

	applied int
	skipped int
	failed  int
	total   int
}

// WithLogger sets the suite logger.
func (s *Suite) WithLogger(log logger.Log) *Suite {
	s.log = log
	return s
}

// Logger returns the underlying logger.
func (s *Suite) Logger() logger.Log {
	return s.log
}

// WithGroups adds groups to the suite and returns the suite.
func (s *Suite) WithGroups(groups ...*Group) *Suite {
	s.groups = append(s.groups, groups...)
	return s
}

// Apply applies the suite.
func (s *Suite) Apply(c *db.Connection) (err error) {
	defer s.writeStats()
	defer func() {
		if r := recover(); r != nil {
			err = exception.New(r)
		}
	}()

	for _, group := range s.groups {
		if err = group.Invoke(s, c); err != nil {
			return
		}
	}
	return
}

func (s *Suite) applyf(group *Group, step Invocable, body string, args ...interface{}) {
	s.applied = s.applied + 1
	s.total = s.total + 1
	s.write(group, step, StatApplied, fmt.Sprintf(body, args...))
}

func (s *Suite) skipf(group *Group, step Invocable, body string, args ...interface{}) {
	s.skipped = s.skipped + 1
	s.total = s.total + 1
	s.write(group, step, StatSkipped, fmt.Sprintf(body, args...))
}

func (s *Suite) errorf(group *Group, step Invocable, body string, args ...interface{}) {
	s.failed = s.failed + 1
	s.total = s.total + 1
	s.write(group, step, StatFailed, fmt.Sprintf(body, args...))
}

func (s *Suite) error(group *Group, step Invocable, err error) error {
	s.failed = s.failed + 1
	s.total = s.total + 1
	s.write(group, step, StatFailed, fmt.Sprintf("%v", err))
	return err
}

func (s *Suite) write(group *Group, step Invocable, result, body string) {
	if s.log == nil {
		return
	}
	var labels []string
	if group != nil && len(group.Label()) > 0 {
		labels = append(labels, group.Label())
	}
	if step != nil && len(step.Label()) > 0 {
		labels = append(labels, step.Label())
	}
	s.log.SyncTrigger(NewEvent(result, body, labels...))
}

func (s *Suite) writeStats() {
	if s.log != nil {
		s.log.SyncTrigger(NewStatsEvent(s.applied, s.skipped, s.failed, s.total))
	}
}
