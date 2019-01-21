package migration

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/blend/go-sdk/logger"
)

const (
	// Flag is a logger event flag.
	Flag logger.Flag = "db.migration"

	// FlagStats is a logger event flag.
	FlagStats logger.Flag = "db.migration.stats"
)

// NewEvent returns a new event.
func NewEvent(result, body string, labels ...string) *Event {
	return &Event{
		EventMeta: logger.NewEventMeta(Flag),
		result:    result,
		body:      body,
		labels:    labels,
	}
}

// Event is a migration logger event.
type Event struct {
	*logger.EventMeta
	result string
	body   string
	labels []string
}

func (e Event) colorizeFixedWidthLeftAligned(tf logger.TextFormatter, text string, color logger.AnsiColor, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	return tf.Colorize(fmt.Sprintf(fixedToken, text), color)
}

// WriteText writes the migration event as text.
func (e Event) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	resultColor := logger.ColorBlue
	switch e.result {
	case "skipped":
		resultColor = logger.ColorYellow
	case "failed":
		resultColor = logger.ColorRed
	}

	if len(e.result) > 0 {
		buf.WriteString(tf.Colorize("--", logger.ColorLightBlack))
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(tf.Colorize(e.result, resultColor))
	}

	if len(e.labels) > 0 {
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(strings.Join(e.labels, " > "))
	}

	if len(e.body) > 0 {
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(tf.Colorize("--", logger.ColorLightBlack))
		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(e.body)
	}
}

// WriteJSON implements logger.JSONWritable.
func (e Event) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		"result": e.result,
		"labels": e.labels,
		"body":   e.body,
	}
}

// NewStatsEvent returns a new stats event.
func NewStatsEvent(applied, skipped, failed, total int) *StatsEvent {
	return &StatsEvent{
		EventMeta: logger.NewEventMeta(FlagStats),
		applied:   applied,
		skipped:   skipped,
		failed:    failed,
		total:     total,
	}
}

// StatsEvent is a migration logger event.
type StatsEvent struct {
	*logger.EventMeta
	applied int
	skipped int
	failed  int
	total   int
}

// WriteText writes the event to a text writer.
func (se StatsEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%s applied %s skipped %s failed %s total",
		tf.Colorize(fmt.Sprintf("%d", se.applied), logger.ColorGreen),
		tf.Colorize(fmt.Sprintf("%d", se.skipped), logger.ColorLightGreen),
		tf.Colorize(fmt.Sprintf("%d", se.failed), logger.ColorRed),
		tf.Colorize(fmt.Sprintf("%d", se.total), logger.ColorLightWhite),
	))
}

// WriteJSON implements logger.JSONWritable.
func (se StatsEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		StatApplied: se.applied,
		StatSkipped: se.skipped,
		StatFailed:  se.failed,
		StatTotal:   se.total,
	}
}
