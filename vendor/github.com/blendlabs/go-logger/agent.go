package logger

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/blendlabs/go-workqueue"
)

var (
	// DefaultAgentQueueWorkers is the number of consumers (goroutines) for the agent work queue.
	DefaultAgentQueueWorkers = 4

	// DefaultAgentQueueLength is the maximum number of items to buffer in the event queue.
	DefaultAgentQueueLength = 1 << 20 // 1mm items
)

var (
	_default     *Agent
	_defaultLock sync.Mutex
)

var (
	// DefaultAgentVerbosity is the default verbosity for a diagnostics agent inited from the environment.
	DefaultAgentVerbosity = NewEventFlagSet(EventFatalError, EventError, EventWebRequest, EventInfo)
)

// Default returnes a default Agent singleton.
func Default() *Agent {
	return _default
}

// SetDefault sets the diagnostics singleton.
func SetDefault(agent *Agent) {
	_defaultLock.Lock()
	defer _defaultLock.Unlock()
	_default = agent
}

// New returns a new diagnostics with a given bitflag verbosity.
func New(events *EventFlagSet) *Agent {
	return &Agent{
		events:         events,
		eventQueue:     newEventQueue(),
		eventListeners: map[EventFlag][]EventListener{},
		debugListeners: []EventListener{},
		writer:         NewWriterWithError(os.Stdout, os.Stderr),
	}
}

// NewWithWriter returns a new diagnostics with a given bitflag verbosity and writer.
func NewWithWriter(events *EventFlagSet, writer *Writer) *Agent {
	return &Agent{
		events:         events,
		eventQueue:     newEventQueue(),
		eventListeners: map[EventFlag][]EventListener{},
		debugListeners: []EventListener{},
		writer:         writer,
	}
}

// NewFromEnvironment returns a new diagnostics with a given bitflag verbosity.
func NewFromEnvironment() *Agent {
	return NewWithWriter(NewEventFlagSetFromEnvironment(), NewWriterFromEnvironment())
}

// All returns a valid agent that fires all events.
func All(writer ...*Writer) *Agent {
	if len(writer) > 0 {
		return NewWithWriter(NewEventFlagSetAll(), writer[0])
	}
	return NewWithWriter(NewEventFlagSetAll(), NewWriterFromEnvironment())
}

// None returns a valid agent that won't fire any events.
func None(writer ...*Writer) *Agent {
	if len(writer) > 0 {
		return NewWithWriter(NewEventFlagSetNone(), writer[0])
	}
	return NewWithWriter(NewEventFlagSetNone(), NewWriterFromEnvironment())
}

// Agent is a handler for various logging events with descendent handlers.
type Agent struct {
	writer             *Writer
	eventsLock         sync.Mutex
	events             *EventFlagSet
	eventListenersLock sync.Mutex
	eventListeners     map[EventFlag][]EventListener
	debugListeners     []EventListener
	eventQueue         *workqueue.Queue
}

// Writer returns the inner Logger for the diagnostics agent.
func (da *Agent) Writer() *Writer {
	return da.writer
}

// EventQueue returns the inner event queue for the agent.
func (da *Agent) EventQueue() *workqueue.Queue {
	return da.eventQueue
}

// Events returns the EventFlagSet
func (da *Agent) Events() *EventFlagSet {
	if da == nil {
		return nil
	}
	return da.events
}

// SetVerbosity sets the agent verbosity synchronously.
func (da *Agent) SetVerbosity(events *EventFlagSet) {
	da.eventsLock.Lock()
	da.events = events
	da.eventsLock.Unlock()
}

// EnableEvent flips the bit flag for a given event.
func (da *Agent) EnableEvent(eventFlag EventFlag) {
	da.eventsLock.Lock()
	da.events.Enable(eventFlag)
	da.eventsLock.Unlock()
}

// DisableEvent flips the bit flag for a given event.
func (da *Agent) DisableEvent(eventFlag EventFlag) {
	da.eventsLock.Lock()
	da.events.Disable(eventFlag)
	da.eventsLock.Unlock()
}

// IsEnabled asserts if a flag value is set or not.
func (da *Agent) IsEnabled(flagValue EventFlag) bool {
	if da == nil {
		return false
	}
	da.eventsLock.Lock()
	enabled := da.events.IsEnabled(flagValue)
	da.eventsLock.Unlock()
	return enabled
}

// HasListener returns if there are registered listener for an event.
func (da *Agent) HasListener(event EventFlag) bool {
	if da == nil {
		return false
	}
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
func (da *Agent) AddEventListener(eventFlag EventFlag, listener EventListener) {
	da.eventListenersLock.Lock()
	da.eventListeners[eventFlag] = append(da.eventListeners[eventFlag], listener)
	da.eventListenersLock.Unlock()
}

// AddDebugListener adds a listener that will fire on *all* events.
func (da *Agent) AddDebugListener(listener EventListener) {
	da.eventListenersLock.Lock()
	da.debugListeners = append(da.debugListeners, listener)
	da.eventListenersLock.Unlock()
}

// RemoveListeners clears *all* listeners for an EventFlag.
func (da *Agent) RemoveListeners(eventFlag EventFlag) {
	delete(da.eventListeners, eventFlag)
}

// OnEvent fires the currently configured event listeners.
func (da *Agent) OnEvent(eventFlag EventFlag, state ...interface{}) {
	if da == nil {
		return
	}
	if da.IsEnabled(eventFlag) && da.HasListener(eventFlag) {
		da.eventQueue.Enqueue(da.triggerListeners, append([]interface{}{TimeNow(), eventFlag}, state...)...)
	}
}

// Infof logs an informational message to the output stream.
func (da *Agent) Infof(format string, args ...interface{}) {
	if da == nil {
		return
	}
	da.WriteEventf(EventInfo, ColorWhite, format, args...)
}

// Debugf logs a debug message to the output stream.
func (da *Agent) Debugf(format string, args ...interface{}) {
	if da == nil {
		return
	}
	da.WriteEventf(EventDebug, ColorLightYellow, format, args...)
}

// Warningf logs a debug message to the output stream.
func (da *Agent) Warningf(format string, args ...interface{}) error {
	if da == nil {
		return nil
	}
	return da.Warning(fmt.Errorf(format, args...))
}

// Warning logs a warning error to std err.
func (da *Agent) Warning(err error) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventWarning, ColorLightYellow, err)
}

// WarningWithReq logs a warning error to std err with a request.
func (da *Agent) WarningWithReq(err error, req *http.Request) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventWarning, ColorLightYellow, err, req)
}

// Errorf writes an event to the log and triggers event listeners.
func (da *Agent) Errorf(format string, args ...interface{}) error {
	if da == nil {
		return nil
	}
	return da.Error(fmt.Errorf(format, args...))
}

// Error logs an error to std err.
func (da *Agent) Error(err error) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventError, ColorRed, err)
}

// ErrorWithReq logs an error to std err with a request.
func (da *Agent) ErrorWithReq(err error, req *http.Request) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventError, ColorRed, err, req)
}

// Fatalf writes an event to the log and triggers event listeners.
func (da *Agent) Fatalf(format string, args ...interface{}) error {
	if da == nil {
		return nil
	}
	return da.Fatal(fmt.Errorf(format, args...))
}

// Fatal logs the result of a panic to std err.
func (da *Agent) Fatal(err error) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventFatalError, ColorRed, err)
}

// FatalWithReq logs the result of a fatal error to std err with a request.
func (da *Agent) FatalWithReq(err error, req *http.Request) error {
	if da == nil {
		return err
	}
	return da.ErrorEventWithState(EventFatalError, ColorRed, err, req)
}

// --------------------------------------------------------------------------------
// meta methods
// --------------------------------------------------------------------------------

// WriteEventf writes to the standard output and triggers events.
func (da *Agent) WriteEventf(event EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if da == nil {
		return
	}
	if da.IsEnabled(event) {
		da.queueWrite(event, ColorLightYellow, format, args...)

		if da.HasListener(event) {
			da.eventQueue.Enqueue(da.triggerListeners, append([]interface{}{TimeNow(), event, format}, args...)...)
		}
	}
}

// WriteErrorEventf writes to the error output and triggers events.
func (da *Agent) WriteErrorEventf(event EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if da == nil {
		return
	}
	if da.IsEnabled(event) {
		da.queueWriteError(event, ColorLightYellow, format, args...)

		if da.HasListener(event) {
			da.eventQueue.Enqueue(da.triggerListeners, append([]interface{}{TimeNow(), event, format}, args...)...)
		}
	}
}

// ErrorEventWithState writes an error and triggers events with a given state.
func (da *Agent) ErrorEventWithState(event EventFlag, color AnsiColorCode, err error, state ...interface{}) error {
	if da == nil {
		return err
	}
	if err != nil {
		if da.IsEnabled(event) {
			da.queueWriteError(event, color, "%+v", err)
			if da.HasListener(event) {
				da.eventQueue.Enqueue(da.triggerListeners, append([]interface{}{TimeNow(), event, err}, state...)...)
			}
		}
	}
	return err
}

// --------------------------------------------------------------------------------
// synchronous methods
// --------------------------------------------------------------------------------

// Sync returns a synchronous agent.
func (da *Agent) Sync() *SyncAgent {
	return &SyncAgent{a: da}
}

// --------------------------------------------------------------------------------
// finalizers
// --------------------------------------------------------------------------------

// Close releases shared resources for the agent.
func (da *Agent) Close() (err error) {
	if da.eventQueue != nil {
		err = da.eventQueue.Close()
		if err != nil {
			return
		}
	}
	if da.writer != nil {
		err = da.writer.Close()
	}
	return
}

// Drain waits for the agent to finish it's queue of events before closing.
func (da *Agent) Drain() error {
	da.SetVerbosity(NewEventFlagSetNone())

	for da.eventQueue.Len() > 0 {
		time.Sleep(time.Millisecond)
	}
	return da.Close()
}

// --------------------------------------------------------------------------------
// internal methods
// --------------------------------------------------------------------------------

// triggerListeners triggers the currently configured event listeners.
func (da *Agent) triggerListeners(actionState ...interface{}) error {
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

	da.eventListenersLock.Lock()
	listeners := da.eventListeners[eventFlag]
	da.eventListenersLock.Unlock()

	for x := 0; x < len(listeners); x++ {
		listener := listeners[x]
		listener(da.writer, timeSource, eventFlag, actionState[2:]...)
	}

	if len(da.debugListeners) > 0 {
		for x := 0; x < len(da.debugListeners); x++ {
			listener := da.debugListeners[x]
			listener(da.writer, timeSource, eventFlag, actionState[2:]...)
		}
	}

	return nil
}

// printf checks an event flag and writes a message with a given color.
func (da *Agent) queueWrite(eventFlag EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if len(format) > 0 {
		da.eventQueue.Enqueue(da.write, append([]interface{}{TimeNow(), eventFlag, color, format}, args...)...)
	}
}

// errorf checks an event flag and writes a message to the error stream (if one is configured) with a given color.
func (da *Agent) queueWriteError(eventFlag EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if len(format) > 0 {
		da.eventQueue.Enqueue(da.writeError, append([]interface{}{TimeNow(), eventFlag, color, format}, args...)...)
	}
}

func (da *Agent) write(actionState ...interface{}) error {
	return da.writeWithOutput(da.writer.PrintfWithTimeSource, actionState...)
}

func (da *Agent) writeError(actionState ...interface{}) error {
	return da.writeWithOutput(da.writer.ErrorfWithTimeSource, actionState...)
}

type loggerOutputWithTimeSource func(ts TimeSource, format string, args ...interface{}) (int64, error)

// writeEventMessage writes an event message.
func (da *Agent) writeWithOutput(output loggerOutputWithTimeSource, actionState ...interface{}) error {
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

	_, err = output(timeSource, "%s %s", da.writer.Colorize(string(eventFlag), labelColor), fmt.Sprintf(format, actionState[4:]...))
	return err
}

func newEventQueue() *workqueue.Queue {
	eq := workqueue.NewWithWorkers(DefaultAgentQueueWorkers)
	eq.SetMaxWorkItems(DefaultAgentQueueLength) //more than this and queuing will block
	eq.Start()
	return eq
}
