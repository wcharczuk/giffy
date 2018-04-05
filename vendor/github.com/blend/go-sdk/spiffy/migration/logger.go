package migration

import (
	"fmt"
	"time"

	"github.com/blend/go-sdk/logger"
)

// NewLogger returns a new logger instance.
func NewLogger(log *logger.Logger) *Logger {
	log.Enable(Flag)
	return &Logger{
		Output: log,
	}
}

// NewLoggerFromEnv returns a new logger instance.
func NewLoggerFromEnv() *Logger {
	log := logger.NewFromEnv()
	log.Enable(Flag)
	return &Logger{
		Output: log,
	}
}

// Logger is a logger for migration steps.
type Logger struct {
	Output *logger.Logger
	Phase  string // `test` or `apply`
	Result string // `apply` or `skipped` or `failed`

	applied int
	skipped int
	failed  int
	total   int
}

// Applyf active actions to the log.
func (l *Logger) Applyf(m Migration, body string, args ...interface{}) error {
	if l == nil {
		return nil
	}

	l.applied = l.applied + 1
	l.total = l.total + 1
	l.Result = "applied"
	l.write(m, fmt.Sprintf(body, args...))
	return nil
}

// Skipf passive actions to the log.
func (l *Logger) Skipf(m Migration, body string, args ...interface{}) error {
	if l == nil {
		return nil
	}
	l.skipped = l.skipped + 1
	l.total = l.total + 1
	l.Result = "skipped"
	l.write(m, fmt.Sprintf(body, args...))
	return nil
}

// Errorf writes errors to the log.
func (l *Logger) Error(m Migration, err error) error {
	if l == nil {
		return err
	}
	l.failed = l.failed + 1
	l.total = l.total + 1
	l.Result = "failed"
	l.write(m, fmt.Sprintf("%v", err.Error()))
	return err
}

// WriteStats writes final stats to output
func (l *Logger) WriteStats() {
	l.Output.SyncTrigger(StatsEvent{
		ts:      time.Now().UTC(),
		applied: l.applied,
		skipped: l.skipped,
		failed:  l.failed,
		total:   l.total,
	})
}

func (l *Logger) colorize(text string, color logger.AnsiColor) string {
	if len(l.Output.Writers()) == 0 {
		return text
	}
	if typed, isTyped := l.Output.Writers()[0].(logger.TextFormatter); isTyped {
		return typed.Colorize(text, color)
	}
	return text
}

func (l *Logger) write(m Migration, body string) {
	l.Output.SyncTrigger(Event{
		ts:     time.Now().UTC(),
		phase:  l.Phase,
		result: l.Result,
		labels: l.labels(m),
		body:   body,
	})
}

func (l *Logger) labels(m Migration) []string {
	labels := []string{m.Label()}
	cursor := m.Parent()
	for cursor != nil {
		if len(cursor.Label()) > 0 {
			labels = append([]string{cursor.Label()}, labels...)
		}
		cursor = cursor.Parent()
	}
	return labels
}
