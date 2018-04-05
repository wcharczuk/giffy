package migration

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/blend/go-sdk/logger"
)

const (
	// Flag is a logger event flag.
	Flag logger.Flag = "db.migration"
)

// Event is a migration logger event.
type Event struct {
	ts     time.Time
	phase  string
	result string
	labels []string
	body   string
}

// Flag returns the logger flag.
func (e Event) Flag() logger.Flag {
	return Flag
}

// Timestamp returns a timestamp.
func (e Event) Timestamp() time.Time {
	return e.ts
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

	buf.WriteString(e.colorizeFixedWidthLeftAligned(tf, e.phase, logger.ColorBlue, 5))
	buf.WriteRune(logger.RuneSpace)
	buf.WriteString(tf.Colorize("--", logger.ColorLightBlack))
	buf.WriteRune(logger.RuneSpace)
	buf.WriteString(e.colorizeFixedWidthLeftAligned(tf, e.result, resultColor, 5))

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
		"phase":  e.phase,
		"result": e.result,
		"labels": e.labels,
		"body":   e.body,
	}
}

// StatsEvent is a migration logger event.
type StatsEvent struct {
	ts      time.Time
	applied int
	skipped int
	failed  int
	total   int
}

// Flag returns the logger flag.
func (se StatsEvent) Flag() logger.Flag {
	return Flag
}

// Timestamp returns a timestamp.
func (se StatsEvent) Timestamp() time.Time {
	return se.ts
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
		"applied": se.applied,
		"skipped": se.skipped,
		"failed":  se.failed,
		"total":   se.total,
	}
}
