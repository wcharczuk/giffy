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

// LoggerEventListener is an event listener for logger events.
type LoggerEventListener func(writer logger.Logger, ts logger.TimeSource, flag logger.EventFlag, query string, elapsed time.Duration, err error, queryLabel string)

// NewLoggerEventListener returns a new listener for diagnostics events.
func NewLoggerEventListener(action LoggerEventListener) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {

		var queryBody = state[0].(string)
		var elapsed = state[1].(time.Duration)

		var err error
		if len(state) > 2 && state[2] != nil {
			err = state[2].(error)
		}

		var queryLabel string
		if len(state) > 3 && state[3] != nil {
			queryLabel = state[3].(string)
		}

		action(writer, ts, eventFlag, queryBody, elapsed, err, queryLabel)
	}
}

// NewPrintStatementListener is a helper listener.
func NewPrintStatementListener() logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		var queryBody = state[0].(string)
		var elapsed = state[1].(time.Duration)

		var err error
		if len(state) > 2 && state[2] != nil {
			err = state[2].(error)
		}

		var queryLabel string
		if len(state) > 3 && state[3] != nil {
			queryLabel = state[3].(string)
		}

		if len(queryLabel) > 0 {
			logger.WriteEventf(writer, ts, eventFlag, logger.ColorLightBlack, "(%v) %s %s", elapsed, queryLabel, queryBody)
		} else {
			logger.WriteEventf(writer, ts, eventFlag, logger.ColorLightBlack, "(%v) %s", elapsed, queryBody)
		}

		if err != nil {
			writer.ErrorfWithTimeSource(ts, "%s", err.Error())
		}
	}
}
