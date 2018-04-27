package migration

import (
	"fmt"

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

// Collector is a results collector.
type Collector struct {
	output  *logger.Logger
	applied int
	skipped int
	failed  int
	total   int
}

// Applyf active actions to the log.
func (c *Collector) Applyf(m Migration, body string, args ...interface{}) {
	if c == nil {
		return
	}
	c.applied = c.applied + 1
	c.total = c.total + 1
	c.write(m, StatApplied, fmt.Sprintf(body, args...))
}

// Skipf passive actions to the log.
func (c *Collector) Skipf(m Migration, body string, args ...interface{}) {
	if c == nil {
		return
	}
	c.skipped = c.skipped + 1
	c.total = c.total + 1
	c.write(m, StatSkipped, fmt.Sprintf(body, args...))
}

// Errorf writes errors to the log.
func (c *Collector) Error(m Migration, err error) error {
	if c == nil {
		return err
	}
	c.failed = c.failed + 1
	c.total = c.total + 1
	c.write(m, StatFailed, fmt.Sprintf("%v", err.Error()))
	return err
}

// WriteStats writes final stats to output
func (c *Collector) WriteStats() {
	if c == nil {
		return
	}
	if c.output == nil {
		return
	}
	c.output.SyncTrigger(NewStatsEvent(c.applied, c.skipped, c.failed, c.total))
}

func (c *Collector) write(m Migration, result, body string) {
	if c == nil {
		return
	}
	if c.output == nil {
		return
	}
	c.output.SyncTrigger(NewEvent(result, body, c.labels(m)...))
}

func (c *Collector) labels(m Migration) []string {
	if c == nil {
		return nil
	}

	var labels []string
	if len(m.Label()) > 0 {
		labels = append(labels, m.Label())
	}

	cursor := m.Parent()
	for cursor != nil {
		if len(cursor.Label()) > 0 {
			labels = append([]string{cursor.Label()}, labels...)
		}
		cursor = cursor.Parent()
	}
	return labels
}
