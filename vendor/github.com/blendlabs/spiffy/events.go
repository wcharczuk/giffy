package spiffy

import (
	"time"

	logger "github.com/blendlabs/go-logger"
)

const (
	// EventFlagExecute is a logger.EventFlag
	EventFlagExecute logger.EventFlag = "db.execute"

	// EventFlagQuery is a logger.EventFlag
	EventFlagQuery logger.EventFlag = "db.query"
)

// EventListener is an event listener for logger events.
type EventListener func(writer *logger.Writer, ts logger.TimeSource, flag logger.EventFlag, query string, elapsed time.Duration, err error, queryLabel string)

// NewEventListener returns a new listener for diagnostics events.
func NewEventListener(action EventListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {

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
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
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
			logger.WriteEventf(writer, ts, eventFlag, logger.ColorLightBlack, "(%v) %s\n%s", elapsed, queryLabel, queryBody)
		} else {
			logger.WriteEventf(writer, ts, eventFlag, logger.ColorLightBlack, "(%v)\n%s", elapsed, queryBody)
		}

		if err != nil {
			writer.ErrorfWithTimeSource(ts, "%s", err.Error())
		}
	}
}
