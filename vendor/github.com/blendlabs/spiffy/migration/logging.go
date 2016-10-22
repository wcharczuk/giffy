package migration

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
)

// NewLogger returns a new logger instance.
func NewLogger() *Logger {
	return &Logger{
		ColorizeOutput: true,
		ShowTimestamp:  true,
		Output:         log.New(os.Stdout, "", 0x0),
	}
}

// NewLoggerFromLog returns a new logger instance from an existing logger..
func NewLoggerFromLog(l *log.Logger) *Logger {
	return &Logger{
		ColorizeOutput: true,
		ShowTimestamp:  true,
		Output:         l,
	}
}

// Logger is a logger for migration steps.
type Logger struct {
	ShowTimestamp  bool
	ColorizeOutput bool

	Output *log.Logger
	Phase  string // `test` or `apply`
	Result string // `apply` or `skipped` or `failed`

	applied int
	skipped int
	failed  int
}

// Applyf active actions to the log.
func (l *Logger) Applyf(m Migration, body string, args ...interface{}) error {
	l.applied = l.applied + 1
	l.Result = "applied"
	l.write(m, util.ColorLightGreen, fmt.Sprintf(body, args...))
	return nil
}

// Skipf passive actions to the log.
func (l *Logger) Skipf(m Migration, body string, args ...interface{}) error {
	l.skipped = l.skipped + 1
	l.Result = "skipped"
	l.write(m, util.ColorGreen, fmt.Sprintf(body, args...))
	return nil
}

// Errorf writes errors to the log.
func (l *Logger) Errorf(m Migration, err error) error {
	l.failed = l.failed + 1
	l.Result = "failed"
	l.write(m, util.ColorRed, fmt.Sprintf("%v", err.Error()))
	return err
}

func (l *Logger) colorize(text string, color util.AnsiColorCode) string {
	if l.ColorizeOutput {
		return color.Apply(text)
	}
	return text
}

func (l *Logger) colorizeFixedWidthLeftAligned(text string, color util.AnsiColorCode, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	fixedMessage := fmt.Sprintf(fixedToken, text)
	if l.ColorizeOutput {
		return color.Apply(fixedMessage)
	}
	return fixedMessage
}

// WriteStats writes final stats to output
func (l *Logger) WriteStats() {
	l.Output.Printf("\n\t%s applied %s skipped %s failed\n\n",
		l.colorize(util.String.IntToString(l.applied), util.ColorGreen),
		l.colorize(util.String.IntToString(l.skipped), util.ColorLightGreen),
		l.colorize(util.String.IntToString(l.failed), util.ColorRed),
	)
}

func (l *Logger) write(m Migration, color util.AnsiColorCode, body string) {
	if l.Output == nil {
		return
	}

	resultColor := util.ColorBlue
	switch l.Result {
	case "skipped":
		resultColor = util.ColorYellow
	case "failed":
		resultColor = util.ColorRed
	}

	var timestamp string
	if l.ShowTimestamp {
		timestamp = l.colorize(time.Now().UTC().Format(time.RFC3339), util.ColorGray) + " "
	}

	l.Output.Printf("%s%s %s %s %s %s %s %s",
		timestamp,
		l.colorize("migrate", util.ColorBlue),
		l.colorizeFixedWidthLeftAligned(l.Phase, util.ColorBlue, 5),
		l.colorize("--", util.ColorLightBlack),
		l.colorizeFixedWidthLeftAligned(l.Result, resultColor, 5),
		l.renderStack(m, color),
		l.colorize("--", util.ColorLightBlack),
		body,
	)
}

func (l *Logger) renderStack(m Migration, color util.AnsiColorCode) string {
	stackSeparator := fmt.Sprintf(" %s ", l.colorize(">", util.ColorLightBlack))
	var renderedStack string
	cursor := m.Parent()
	for cursor != nil {
		if len(cursor.Label()) > 0 {
			renderedStack = stackSeparator + cursor.Label() + renderedStack
		}
		cursor = cursor.Parent()
	}
	return strings.TrimPrefix(renderedStack, " ")
}
