package migration

import (
	"fmt"
	"log"
	"os"

	"github.com/blendlabs/go-util"
)

// NewLogger returns a new logger instance.
func NewLogger() *log.Logger {
	return log.New(os.Stdout, util.Color("migrate ", util.ColorBlue), 0x0)
}

func logActionActive(l *log.Logger, actionName, body string, args ...interface{}) {
	if l != nil {
		l.Printf("%s %s", util.ColorFixedWidthLeftAligned(actionName, util.ColorGreen, 15), fmt.Sprintf(body, args...))
	}
}

func logActionPassive(l *log.Logger, actionName, body string, args ...interface{}) {
	if l != nil {
		l.Printf("%s %s", util.ColorFixedWidthLeftAligned(actionName, util.ColorGreen, 15), fmt.Sprintf(body, args...))
	}
}

func logError(l *log.Logger, err error) {
	if l != nil {
		l.Printf("%s %v", util.ColorFixedWidthLeftAligned("error", util.ColorRed, 15), err)
	}
}

func setLoggerPhase(l *log.Logger, phase, name string) {
	if l != nil {
		if !util.IsEmpty(name) {
			labelAndPhase := util.ColorFixedWidthLeftAligned(fmt.Sprintf("migrate (%s)", phase), util.ColorBlue, 16)
			migrationName := util.ColorFixedWidthLeftAligned(name, util.ColorReset, 18)
			l.SetPrefix(fmt.Sprintf("%s %s", labelAndPhase, migrationName))
		} else if !util.IsEmpty(phase) {
			l.SetPrefix(util.ColorFixedWidthLeftAligned(fmt.Sprintf("migrate (%s)", phase), util.ColorBlue, 16))
		} else {
			l.SetPrefix(util.ColorFixedWidthLeftAligned("migrate", util.ColorBlue, 16))
		}
	}
}
