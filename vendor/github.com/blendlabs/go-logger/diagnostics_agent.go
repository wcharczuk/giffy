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
	DefaultDiagnosticsAgentQueueLength = 1 << 20
)

var (
	_diagnosticsAgent     *DiagnosticsAgent
	_diagnosticsAgentLock sync.Mutex
)

// InitializeDiagnostics initializes the Diagnostics() agent with a given verbosity
// and optionally a targeted writer (only the first variadic writer will be used).
func InitializeDiagnostics(verbosity uint64, writers ...Logger) {
	_diagnosticsAgentLock.Lock()
	defer _diagnosticsAgentLock.Unlock()

	_diagnosticsAgent = NewDiagnosticsAgent(verbosity, writers...)
}

// InitializeDiagnosticsFromEnvironment initializes the Diagnostics() agent with a given verbosity
// and optionally a targeted writer (only the first variadic writer will be used).
func InitializeDiagnosticsFromEnvironment() error {
	_diagnosticsAgentLock.Lock()
	defer _diagnosticsAgentLock.Unlock()

	agent, err := NewDiagnosticsAgentFromEnvironment()
	if err != nil {
		return err
	}
	_diagnosticsAgent = agent
	return nil
}

// Diagnostics returnes a default DiagnosticsAgent singleton.
func Diagnostics() *DiagnosticsAgent {
	if _diagnosticsAgent == nil {
		_diagnosticsAgentLock.Lock()
		defer _diagnosticsAgentLock.Unlock()
		if _diagnosticsAgent == nil {
			_diagnosticsAgent = NewDiagnosticsAgent(EventNone) // do this so .Diagnostics() calls don't panic
		}
	}
	return _diagnosticsAgent
}

// NewDiagnosticsAgent returns a new diagnostics with a given bitflag verbosity.
func NewDiagnosticsAgent(verbosity uint64, writers ...Logger) *DiagnosticsAgent {
	diag := &DiagnosticsAgent{
		verbosity:  verbosity,
		eventQueue: workQueue.NewQueueWithWorkers(DefaultDiagnosticsAgentQueueWorkers),
	}
	diag.eventQueue.UseSynchronousDispatch()                            //dispatch items in order
	diag.eventQueue.SetMaxWorkItems(DefaultDiagnosticsAgentQueueLength) //more than this and queuing will block
	diag.eventQueue.Start()

	if len(writers) > 0 {
		diag.writer = writers[0]
	} else {
		diag.writer = NewLogWriter(os.Stdout, os.Stderr)
	}
	return diag
}

// NewDiagnosticsAgentFromEnvironment returns a new diagnostics with a given bitflag verbosity.
func NewDiagnosticsAgentFromEnvironment() (*DiagnosticsAgent, error) {
	var err error
	envEventFlag := os.Getenv("LOG_VERBOSITY")
	eventFlag := EventFlagCombine(EventFatalError, EventError, EventRequestComplete, EventInfo)
	if len(envEventFlag) > 0 {
		eventFlag, err = ParseEventFlagNameSet(envEventFlag)
		if err != nil {
			return nil, err
		}
	}
	diag := &DiagnosticsAgent{
		verbosity:  eventFlag,
		eventQueue: workQueue.NewQueueWithWorkers(DefaultDiagnosticsAgentQueueWorkers),
		writer:     NewLogWriterFromEnvironment(),
	}
	diag.eventQueue.UseSynchronousDispatch()                            //dispatch items in order
	diag.eventQueue.SetMaxWorkItems(DefaultDiagnosticsAgentQueueLength) //more than this and queuing will block
	diag.eventQueue.Start()

	return diag, nil
}

// DiagnosticsAgent is a handler for various logging events with descendent handlers.
type DiagnosticsAgent struct {
	writer         Logger
	verbosity      uint64
	eventListeners map[uint64][]EventListener
	eventQueue     *workQueue.Queue
}

// Label returns the label for the diagnostics agent.
func (da *DiagnosticsAgent) Label() string {
	return da.writer.Label()
}

// SetLabel sets the logging label for the diagnostics agent.
func (da *DiagnosticsAgent) SetLabel(label string) {
	da.writer.SetLabel(label)
}

// AddEventListener adds a listener for errors.
func (da *DiagnosticsAgent) AddEventListener(eventFlag uint64, listener EventListener) {
	if da.eventListeners == nil {
		da.eventListeners = map[uint64][]EventListener{}
	}
	da.eventListeners[eventFlag] = append(da.eventListeners[eventFlag], listener)
}

// OnEvent fires the currently configured event listeners.
func (da *DiagnosticsAgent) OnEvent(eventFlag uint64, state ...interface{}) {
	if da.CheckVerbosity(eventFlag) {
		if da.CheckHasHandler(eventFlag) {
			da.eventQueue.Enqueue(da.fireEvent, append([]interface{}{Now(), eventFlag}, state...)...)
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

// QueueLen returns the length of the queue.
func (da *DiagnosticsAgent) QueueLen() int {
	return da.eventQueue.Len()
}

// Verbosity sets the agent verbosity synchronously.
func (da *DiagnosticsAgent) Verbosity() uint64 {
	return da.verbosity
}

// SetVerbosity sets the agent verbosity synchronously.
func (da *DiagnosticsAgent) SetVerbosity(verbosity uint64) {
	da.verbosity = verbosity
}

// CheckVerbosity asserts if a flag value is set or not.
func (da *DiagnosticsAgent) CheckVerbosity(flagValue uint64) bool {
	return EventFlagAny(da.verbosity, flagValue)
}

// CheckHasHandler returns if there are registered handlers for an event.
func (da *DiagnosticsAgent) CheckHasHandler(event uint64) bool {
	_, hasHandler := da.eventListeners[event]
	return hasHandler
}

// Eventf checks an event flag and writes a message with a given label and color.
func (da *DiagnosticsAgent) Eventf(eventFlag uint64, label string, labelColor AnsiColorCode, format string, args ...interface{}) {
	if da.CheckVerbosity(eventFlag) && len(format) > 0 {
		defer da.OnEvent(eventFlag)
		da.eventQueue.Enqueue(da.writeEventMessage, append([]interface{}{Now(), label, labelColor, format}, args...)...)
	}
}

// ErrorEventf checks an event flag and writes a message with a given label and color.
func (da *DiagnosticsAgent) ErrorEventf(eventFlag uint64, label string, labelColor AnsiColorCode, format string, args ...interface{}) {
	if da.CheckVerbosity(eventFlag) && len(format) > 0 {
		defer da.OnEvent(eventFlag)
		da.eventQueue.Enqueue(da.writeErrorEventMessage, append([]interface{}{Now(), label, labelColor, format}, args...)...)
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
	label, err := stateAsString(actionState[1])
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

	output(timeSource, "%s %s", da.writer.Colorize(label, labelColor), fmt.Sprintf(format, actionState[4:]...))
	return nil
}

// Infof logs an informational message to the output stream.
func (da *DiagnosticsAgent) Infof(format string, args ...interface{}) {
	da.Eventf(EventInfo, "Info", ColorWhite, format, args...)
}

// Debugf logs a debug message to the output stream.
func (da *DiagnosticsAgent) Debugf(format string, args ...interface{}) {
	da.Eventf(EventDebug, "Debug", ColorLightYellow, format, args...)
}

// DebugDump dumps an object and fires a debug event.
func (da *DiagnosticsAgent) DebugDump(object interface{}) {
	da.Eventf(EventDebug, "Debug Dump", ColorLightYellow, "%v", object)
}

// Warningf logs a debug message to the output stream.
func (da *DiagnosticsAgent) Warningf(format string, args ...interface{}) {
	da.ErrorEventf(EventWarning, "Warning", ColorYellow, format, args...)
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
	da.ErrorEventf(EventError, "Error", ColorRed, format, args...)
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
	da.ErrorEventf(EventFatalError, "Fatal Error", ColorRed, format, args...)
}

// Fatal logs the result of a panic to std err.
func (da *DiagnosticsAgent) Fatal(err interface{}) {
	if err != nil {
		da.Fatalf("%v", err)
	}
}