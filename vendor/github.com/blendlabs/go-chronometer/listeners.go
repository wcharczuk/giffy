package chronometer

import (
	"time"

	"github.com/blendlabs/go-logger"
)

const (
	// EventTask is a logger diagnostics event for task completions.
	EventTask logger.EventFlag = "chronometer.task"
	// EventTaskComplete is a logger diagnostics event for task completions.
	EventTaskComplete logger.EventFlag = "chronometer.task.complete"
)

// TaskListener is a listener for task complete events.
type TaskListener func(w *logger.Writer, ts logger.TimeSource, taskName string)

// NewTaskListener returns a new event listener for task events.
func NewTaskListener(listener TaskListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) > 0 {
			listener(writer, ts, state[0].(string))
		}
	}
}

// TaskCompleteListener is a listener for task complete events.
type TaskCompleteListener func(w *logger.Writer, ts logger.TimeSource, taskName string, elapsed time.Duration, err error)

// NewTaskCompleteListener returns a new event listener for task events.
func NewTaskCompleteListener(listener TaskCompleteListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) > 2 {
			if state[2] == nil {
				listener(writer, ts, state[0].(string), state[1].(time.Duration), nil)
			} else {
				listener(writer, ts, state[0].(string), state[1].(time.Duration), state[2].(error))
			}
		}
	}
}
