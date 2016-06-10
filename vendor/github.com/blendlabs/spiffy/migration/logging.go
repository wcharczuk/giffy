package migration

import (
	"fmt"
	"log"
	"os"

	"github.com/blendlabs/go-util"
)

// NewLogger returns a new logger instance.
func NewLogger() *Logger {
	return &Logger{
		Output: log.New(os.Stdout, "", 0x0),
	}
}

// NewLoggerFromLog returns a new logger instance from an existing logger..
func NewLoggerFromLog(l *log.Logger) *Logger {
	return &Logger{
		Output: l,
	}
}

// Logger is a logger for migration steps.
type Logger struct {
	Output *log.Logger
	Phase  string // `test` or `apply`

	applied int
	skipped int
	failed  int
}

// Applyf active actions to the log.
func (l *Logger) Applyf(stack []string, body string, args ...interface{}) error {
	l.applied = l.applied + 1
	l.write(stack, util.ColorLightGreen, fmt.Sprintf(body, args...))
	return nil
}

// Skipf passive actions to the log.
func (l *Logger) Skipf(stack []string, body string, args ...interface{}) error {
	l.skipped = l.skipped + 1
	l.write(stack, util.ColorGreen, fmt.Sprintf(body, args...))
	return nil
}

// Errorf writes errors to the log.
func (l *Logger) Errorf(stack []string, err error) error {
	l.failed = l.failed + 1
	l.write(stack, util.ColorRed, fmt.Sprintf("%v", err.Error()))
	return err
}

// WriteStats writes final stats to output
func (l *Logger) WriteStats() {
	l.Output.Printf("\n\t%s applied %s skipped %s failed\n\n",
		util.Color(util.IntToString(l.applied), util.ColorGreen),
		util.Color(util.IntToString(l.skipped), util.ColorLightGreen),
		util.Color(util.IntToString(l.failed), util.ColorRed),
	)
}

func (l *Logger) write(stack []string, color, body string) {
	l.Output.Printf("%s %s %s %s %s %s",
		util.Color("migrate", util.ColorBlue),
		util.ColorFixedWidthLeftAligned(l.Phase, util.ColorBlue, 5),
		util.Color("--", util.ColorLightBlack),
		l.renderStack(stack, color),
		util.Color("--", util.ColorLightBlack),
		body,
	)
}

func (l *Logger) renderStack(stack []string, color string) string {
	stackSeparator := util.Color(" > ", util.ColorLightBlack)
	var renderedStack string
	for index, stackElement := range stack {
		if len(stackElement) == 0 {
			continue
		}

		if index < len(stack)-1 {
			renderedStack = renderedStack + util.Color(stackElement, color)
			renderedStack = renderedStack + stackSeparator
		} else {
			renderedStack = renderedStack + stackElement
		}
	}
	return renderedStack
}
