package migration

import (
	"fmt"
	"log"
	"os"
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

func (l *Logger) colorize(text, color string) string {
	if l.ColorizeOutput {
		return util.Color(text, color)
	}
	return text
}

func (l *Logger) colorizeFixedWidthLeftAligned(text, color string, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	fixedMessage := fmt.Sprintf(fixedToken, text)
	if l.ColorizeOutput {
		return fmt.Sprintf("%s%s%s", util.AnsiEscapeCode(color), fixedMessage, util.AnsiEscapeCode(util.ColorReset))
	}
	return fixedMessage
}

// WriteStats writes final stats to output
func (l *Logger) WriteStats() {
	l.Output.Printf("\n\t%s applied %s skipped %s failed\n\n",
		l.colorize(util.IntToString(l.applied), util.ColorGreen),
		l.colorize(util.IntToString(l.skipped), util.ColorLightGreen),
		l.colorize(util.IntToString(l.failed), util.ColorRed),
	)
}

func (l *Logger) write(m Migration, color, body string) {
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

func (l *Logger) renderStack(m Migration, color string) string {
	stackSeparator := l.colorize(" > ", util.ColorLightBlack)
	var renderedStack string
	cursor := m.Parent()
	for cursor != nil {
		if len(cursor.Label()) > 0 {
			renderedStack = stackSeparator + cursor.Label() + renderedStack
		}
		cursor = cursor.Parent()
	}
	return renderedStack
}
