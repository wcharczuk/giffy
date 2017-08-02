package logger

import (
	"fmt"
	"net/http"
	"os"
)

// SyncAgent is an agent that fires events synchronously.
// It wraps a regular agent.
type SyncAgent struct {
	a *Agent
}

// Agent returns the underlying agent.
func (sa *SyncAgent) Agent() *Agent {
	return sa.a
}

// Writer returns the underlying writer.
func (sa *SyncAgent) Writer() *Writer {
	return sa.a.Writer()
}

// Infof logs an informational message to the output stream.
func (sa *SyncAgent) Infof(format string, args ...interface{}) {
	if sa == nil {
		return
	}
	sa.WriteEventf(EventInfo, ColorLightWhite, format, args...)
}

// Debugf logs a debug message to the output stream.
func (sa *SyncAgent) Debugf(format string, args ...interface{}) {
	if sa == nil {
		return
	}
	sa.WriteEventf(EventDebug, ColorLightYellow, format, args...)
}

// Warningf logs a debug message to the output stream.
func (sa *SyncAgent) Warningf(format string, args ...interface{}) error {
	if sa == nil {
		return nil
	}
	return sa.Warning(fmt.Errorf(format, args...))
}

// Warning logs a warning error to std err.
func (sa *SyncAgent) Warning(err error) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventWarning, ColorLightYellow, err)
}

// WarningWithReq logs a warning error to std err with a request.
func (sa *SyncAgent) WarningWithReq(err error, req *http.Request) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventWarning, ColorLightYellow, err, req)
}

// Errorf writes an event to the log and triggers event listeners.
func (sa *SyncAgent) Errorf(format string, args ...interface{}) error {
	if sa == nil {
		return nil
	}
	return sa.Error(fmt.Errorf(format, args...))
}

// Error logs an error to std err.
func (sa *SyncAgent) Error(err error) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventError, ColorRed, err)
}

// ErrorWithReq logs an error to std err with a request.
func (sa *SyncAgent) ErrorWithReq(err error, req *http.Request) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventError, ColorRed, err, req)
}

// Fatalf writes an event to the log and triggers event listeners.
func (sa *SyncAgent) Fatalf(format string, args ...interface{}) error {
	if sa == nil {
		return nil
	}
	return sa.Fatal(fmt.Errorf(format, args...))
}

// Fatal logs the result of a panic to std err.
func (sa *SyncAgent) Fatal(err error) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventFatalError, ColorRed, err)
}

// FatalWithReq logs the result of a fatal error to std err with a request.
func (sa *SyncAgent) FatalWithReq(err error, req *http.Request) error {
	if sa == nil {
		return err
	}
	return sa.ErrorEventWithState(EventFatalError, ColorRed, err, req)
}

// FatalExit logs the result of a fatal error to std err and calls `exit(1)`.
// NOTE: this terminates the program.
func (sa *SyncAgent) FatalExit(err error) {
	if sa == nil {
		os.Exit(1)
	}

	sa.ErrorEventWithState(EventFatalError, ColorRed, err)
	os.Exit(1)
}

// WriteEventf writes to the standard output and triggers events.
func (sa *SyncAgent) WriteEventf(event EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if sa == nil {
		return
	}
	if sa.a == nil {
		return
	}
	if sa.a.IsEnabled(event) {
		sa.a.write(append([]interface{}{TimeNow(), event, color, format}, args...)...)

		if sa.a.HasListener(event) {
			sa.a.triggerListeners(append([]interface{}{TimeNow(), event, format}, args...)...)
		}
	}
}

// WriteErrorEventf writes to the error output and triggers events.
func (sa *SyncAgent) WriteErrorEventf(event EventFlag, color AnsiColorCode, format string, args ...interface{}) {
	if sa == nil {
		return
	}
	if sa.a == nil {
		return
	}
	if sa.a.IsEnabled(event) {
		sa.a.writeError(append([]interface{}{TimeNow(), event, color, format}, args...)...)

		if sa.a.HasListener(event) {
			sa.a.triggerListeners(append([]interface{}{TimeNow(), event, format}, args...)...)
		}
	}
}

// ErrorEventWithState writes an error and triggers events with a given state.
func (sa *SyncAgent) ErrorEventWithState(event EventFlag, color AnsiColorCode, err error, state ...interface{}) error {
	if sa == nil {
		return err
	}
	if sa.a == nil {
		return err
	}
	if err != nil {
		if sa.a.IsEnabled(event) {
			sa.a.writeError(TimeNow(), event, color, "%+v", err)
			if sa.a.HasListener(event) {
				sa.a.triggerListeners(append([]interface{}{TimeNow(), event, err}, state...)...)
			}
		}
	}
	return err
}

// OnEvent fires the currently configured event listeners.
func (sa *SyncAgent) OnEvent(eventFlag EventFlag, state ...interface{}) {
	if sa == nil {
		return
	}
	if sa.a == nil {
		return
	}
	if sa.a.IsEnabled(eventFlag) && sa.a.HasListener(eventFlag) {
		sa.a.triggerListeners(append([]interface{}{TimeNow(), eventFlag}, state...)...)
	}
}
