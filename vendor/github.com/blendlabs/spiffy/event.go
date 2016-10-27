package spiffy

import (
	"time"

	logger "github.com/blendlabs/go-logger"
)

const (
	// EventFlagExecute is a logger.EventFlag
	EventFlagExecute logger.EventFlag = "spiffy.execute"

	// EventFlagQuery is a logger.EventFlag
	EventFlagQuery logger.EventFlag = "spiffy.query"
)

// NewLoggerEventListener returns a new listener for diagnostics events.
func NewLoggerEventListener(action func(writer logger.Logger, ts logger.TimeSource, flag logger.EventFlag, query string, elapsed time.Duration, err error)) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if state[2] != nil {
			action(writer, ts, eventFlag, state[0].(string), state[1].(time.Duration), state[2].(error))
		} else {
			action(writer, ts, eventFlag, state[0].(string), state[1].(time.Duration), nil)
		}
	}
}
