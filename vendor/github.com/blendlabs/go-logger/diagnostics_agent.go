package logger

import (
	"fmt"
	"os"
	"sync"

	workQueue "github.com/blendlabs/go-workqueue"
)

var (
	// DefaultDiagnosticsAgentQueueWorkers is the number of consumers
	// for the diagnostics agent work queue.
	DefaultDiagnosticsAgentQueueWorkers = 1

	// DefaultDiagnosticsAgentQueueLength is the maximum number of items to buffer in the event queue.
	DefaultDiagnosticsAgentQueueLength = 1 << 10 // 1024
)

var (
	_diagnosticsAgent     *DiagnosticsAgent
	_diagnosticsAgentLock sync.Mutex
)

var (
	// DefaultDiagnosticsAgentVerbosity is the default verbosity for a diagnostics agent inited from the environment.
	DefaultDiagnosticsAgentVerbosity = NewEventFlagSetWithEvents(EventFatalError, EventError, EventWebRequest, EventInfo)
)

// Diagnostics returnes a default DiagnosticsAgent singleton.
func Diagnostics() *DiagnosticsAgent {
	return _diagnosticsAgent
}

// SetDiagnostics sets the diagnostics singleton.
func SetDiagnostics(diagnostics *DiagnosticsAgent) {
	_diagnosticsAgentLock.Lock()
	defer _diagnosticsAgentLock.Unlock()
	_diagnosticsAgent = diagnostics
}

func newDiagnosticsEventQueue() *workQueue.Queue {
	eq := workQueue.NewQueueWithWorkers(DefaultDiagnosticsAgentQueueWorkers)
	eq.UseSynchronousDispatch()                            //dispatch items in order
	eq.SetMaxWorkItems(DefaultDiagnosticsAgentQueueLength) //more than this and queuing will block
	return eq
}

// NewDiagnosticsAgent returns a new diagnostics with a given bitflag verbosity.
func NewDiagnosticsAgent(events *EventFlagSet, optionalWriter ...Logger) *DiagnosticsAgent {
	diag := &DiagnosticsAgent{
		events:         events,
		eventQueue:     newDiagnosticsEventQueue(),
		eventListeners: map[EventFlag][]EventListener{},
	}

	if len(optionalWriter) > 0 {
		diag.writer = optionalWriter[0]
	} else {
		diag.writer = NewLogWriter(os.Stdout, os.Stderr)
	}
	return diag
}

// NewDiagnosticsAgentFromEnvironment returns a new diagnostics with a given bitflag verbosity.
func NewDiagnosticsAgentFromEnvironment() *DiagnosticsAgent {
	return NewDiagnosticsAgent(NewEventFlagSetFromEnvironment(), NewLogWriterFromEnvironment())
}

// DiagnosticsAgent is a handler for various logging events with descendent handlers.
type DiagnosticsAgent struct {
	writer         Logger
	events         *EventFlagSet
	eventListeners map[EventFlag][]EventListener
	eventQueue     *workQueue.Queue
}

// Writer returns the inner Logger for the diagnostics agent.
func (da *DiagnosticsAgent) Writer() Logger {
	return da.writer
}

// EventQueue returns the inner event queue for the agent.
func (da *DiagnosticsAgent) EventQueue() *workQueue.Queue {
	return da.eventQueue
}

// Events returns the EventFlagSet
func (da *DiagnosticsAgent) Events() *EventFlagSet {
	return da.events
}

// SetVerbosity sets the agent verbosity synchronously.
func (da *DiagnosticsAgent) SetVerbosity(events *EventFlagSet) {
	da.events = events
}

// EnableEvent flips the bit flag for a given event.
func (da *DiagnosticsAgent) EnableEvent(eventFlag EventFlag) {
	da.events.Enable(eventFlag)
}

// DisableEvent flips the bit flag for a given event.
func (da *DiagnosticsAgent) DisableEvent(eventFlag EventFlag) {
	da.events.Disable(eventFlag)
}

// IsEnabled asserts if a flag value is set or not.
func (da *DiagnosticsAgent) IsEnabled(flagValue EventFlag) bool {
	return da.events.IsEnabled(flagValue)
}

// HasListener returns if there are registered listener for an event.
func (da *DiagnosticsAgent) HasListener(event EventFlag) bool {
	if da.eventListeners == nil {
		return false
	}
	listeners, hasHandler := da.eventListeners[event]
	if !hasHandler {
		return false
	}
	return len(listeners) > 0
}

// AddEventListener adds a listener for errors.
func (da *DiagnosticsAgent) AddEventListener(eventFlag EventFlag, listener EventListener) {
	da.eventListeners[eventFlag] = append(da.eventListeners[eventFlag], listener)
}

// RemoveListeners clears *all* listeners for an EventFlag.
func (da *DiagnosticsAgent) RemoveListeners(eventFlag EventFlag) {
	delete(da.eventListeners, eventFlag)
}

// OnEvent fires the currently configured event listeners.
func (da *DiagnosticsAgent) OnEvent(eventFlag EventFlag, state ...interface{}) {
	if da.IsEnabled(eventFlag) {
		if da.HasListener(eventFlag) {
			if !da.eventQueue.Running() {
				da.eventQueue.Start()
			}

			da.eventQueue.Enqueue(da.fireEvent, append([]interface{}{TimeNow(), eventFlag}, state...)...)
		}
	}
}

// OnEvent fires the currently configured event listeners.
func (da *DiagnosticsAgent) fireEvent(actionState ...interface{}) error {
	if len(actionState) < 2 {
		return nil
	}

	timeSource, err := stateAsTimeSource(actionState[0])
	if err != nil {
		return err
	}

	eventFlag, err := stateAsEventFlag(actionState[1])
	if err != nil {
		return err
	}

	listeners := da.eventListeners[eventFlag]
	for x := 0; x < len(listeners); x++ {
		listener := listeners[x]
		listener(da.writer, timeSource, eventFlag, actionState[2:]...)
	}

	return nil
}

// Eventf checks an event flag and writes a message with a given color.
func (da *DiagnosticsAgent) Eventf(eventFlag EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if da.IsEnabled(eventFlag) && len(format) > 0 {
		defer da.OnEvent(eventFlag)

		if !da.eventQueue.Running() {
			da.eventQueue.Start()
		}
		da.eventQueue.Enqueue(da.writeEventMessage, append([]interface{}{TimeNow(), eventFlag, color, format}, args...)...)
	}
}

// ErrorEventf checks an event flag and writes a message to the error stream (if one is configured) with a given color.
func (da *DiagnosticsAgent) ErrorEventf(eventFlag EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if da.IsEnabled(eventFlag) && len(format) > 0 {
		defer da.OnEvent(eventFlag)

		if !da.eventQueue.Running() {
			da.eventQueue.Start()
		}
		da.eventQueue.Enqueue(da.writeErrorEventMessage, append([]interface{}{TimeNow(), eventFlag, color, format}, args...)...)
	}
}

func (da *DiagnosticsAgent) writeEventMessage(actionState ...interface{}) error {
	return da.writeEventMessageWithOutput(da.writer.PrintfWithTimeSource, actionState...)
}

func (da *DiagnosticsAgent) writeErrorEventMessage(actionState ...interface{}) error {
	return da.writeEventMessageWithOutput(da.writer.ErrorfWithTimeSource, actionState...)
}

// writeEventMessage writes an event message.
func (da *DiagnosticsAgent) writeEventMessageWithOutput(output loggerOutputWithTimeSource, actionState ...interface{}) error {
	if len(actionState) < 4 {
		return nil
	}

	timeSource, err := stateAsTimeSource(actionState[0])
	if err != nil {
		return err
	}

	eventFlag, err := stateAsEventFlag(actionState[1])
	if err != nil {
		return err
	}

	labelColor, err := stateAsAnsiColorCode(actionState[2])
	if err != nil {
		return err
	}

	format, err := stateAsString(actionState[3])
	if err != nil {
		return err
	}

	output(timeSource, "%s %s", da.writer.Colorize(string(eventFlag), labelColor), fmt.Sprintf(format, actionState[4:]...))
	return nil
}

// Infof logs an informational message to the output stream.
func (da *DiagnosticsAgent) Infof(format string, args ...interface{}) {
	da.Eventf(EventInfo, ColorWhite, format, args...)
}

// Debugf logs a debug message to the output stream.
func (da *DiagnosticsAgent) Debugf(format string, args ...interface{}) {
	da.Eventf(EventDebug, ColorLightYellow, format, args...)
}

// DebugDump dumps an object and fires a debug event.
func (da *DiagnosticsAgent) DebugDump(object interface{}) {
	da.Eventf(EventDebug, ColorLightYellow, "%v", object)
}

// Warningf logs a debug message to the output stream.
func (da *DiagnosticsAgent) Warningf(format string, args ...interface{}) {
	da.ErrorEventf(EventWarning, ColorYellow, format, args...)
}

// Warning logs a warning error to std err.
func (da *DiagnosticsAgent) Warning(err error) error {
	if err != nil {
		da.Warningf(err.Error())
	}
	return err
}

// Errorf writes an event to the log and triggers event listeners.
func (da *DiagnosticsAgent) Errorf(format string, args ...interface{}) {
	da.ErrorEventf(EventError, ColorRed, format, args...)
}

// Fatal logs an error to std err.
func (da *DiagnosticsAgent) Error(err error) error {
	if err != nil {
		da.Errorf(err.Error())
	}
	return err
}

// Fatalf writes an event to the log and triggers event listeners.
func (da *DiagnosticsAgent) Fatalf(format string, args ...interface{}) {
	da.ErrorEventf(EventFatalError, ColorRed, format, args...)
}

// Fatal logs the result of a panic to std err.
func (da *DiagnosticsAgent) Fatal(err interface{}) {
	if err != nil {
		da.Fatalf("%v", err)
	}
}

// Close releases shared resources for the agent.
func (da *DiagnosticsAgent) Close() error {
	return da.eventQueue.Drain()
}
