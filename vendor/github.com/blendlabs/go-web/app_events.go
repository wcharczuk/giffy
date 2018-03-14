package web

import (
	"bytes"
	"fmt"
	"time"

	logger "github.com/blendlabs/go-logger"
)

const (
	// FlagAppStart fires when the app is starting.
	FlagAppStart logger.Flag = "web.app.start"

	// FlagAppStartComplete fires after the app has started.
	FlagAppStartComplete logger.Flag = "web.app.start.complete"

	// FlagAppExit fires when an app exits.
	FlagAppExit logger.Flag = "web.app.exit"
)

// NewAppStartEvent creates a new app start event.
func NewAppStartEvent(app *App) AppStartEvent {
	return AppStartEvent{
		ts:  time.Now().UTC(),
		app: app,
	}
}

// NewAppStartEventListener returns a new app start event listener.
func NewAppStartEventListener(listener func(me AppStartEvent)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(AppStartEvent); isTyped {
			listener(typed)
		}
	}
}

// AppStartEvent is an event.
type AppStartEvent struct {
	ts  time.Time
	app *App
}

// Flag returns the logger flag.
func (ae AppStartEvent) Flag() logger.Flag {
	return FlagAppStart
}

// Timestamp returns the timestamp for a
func (ae AppStartEvent) Timestamp() time.Time {
	return ae.ts
}

// App returns the app reference.
func (ae AppStartEvent) App() *App {
	return ae.app
}

// String implements fmt.Stringer
func (ae AppStartEvent) String() string {
	return "started"
}

// WriteText implements logger.TextWritable.
func (ae AppStartEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(ae.String())
}

// WriteJSON implements logger.JSONWritable.
func (ae AppStartEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		logger.JSONFieldMessage: ae.String(),
	}
}

// NewAppStartCompleteEvent creates a new app start complete event.
func NewAppStartCompleteEvent(app *App, elapsed time.Duration, err error) AppStartCompleteEvent {
	return AppStartCompleteEvent{
		ts:      time.Now().UTC(),
		app:     app,
		elapsed: elapsed,
		err:     err,
	}
}

// NewAppStartCompleteEventListener returns a new app start complete event listener.
func NewAppStartCompleteEventListener(listener func(me AppStartCompleteEvent)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(AppStartCompleteEvent); isTyped {
			listener(typed)
		}
	}
}

// AppStartCompleteEvent is an event.
type AppStartCompleteEvent struct {
	ts      time.Time
	app     *App
	elapsed time.Duration
	err     error
}

// Flag returns the logger flag.
func (ae AppStartCompleteEvent) Flag() logger.Flag {
	return FlagAppStartComplete
}

// Timestamp returns the timestamp for a
func (ae AppStartCompleteEvent) Timestamp() time.Time {
	return ae.ts
}

// App returns the app reference.
func (ae AppStartCompleteEvent) App() *App {
	return ae.app
}

// Elapsed returns the elapsed time.
func (ae AppStartCompleteEvent) Elapsed() time.Duration {
	return ae.elapsed
}

// Err retruns an error.
func (ae AppStartCompleteEvent) Err() error {
	return ae.err
}

// String implements fmt.Stringer.
func (ae AppStartCompleteEvent) String() string {
	if ae.err != nil {
		return fmt.Sprintf("failed (%v)", ae.elapsed)
	}
	return fmt.Sprintf("complete (%v)", ae.elapsed)
}

// WriteText implements logger.TextWritable.
func (ae AppStartCompleteEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(ae.String())
}

// WriteJSON implements logger.JSONWritable.
func (ae AppStartCompleteEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		logger.JSONFieldErr:     ae.Err(),
		logger.JSONFieldElapsed: logger.Milliseconds(ae.Elapsed()),
		logger.JSONFieldMessage: ae.String(),
	}
}

// NewAppExitEvent creates a new app exit event.
func NewAppExitEvent(app *App, err error) AppExitEvent {
	return AppExitEvent{
		ts:  time.Now().UTC(),
		app: app,
		err: err,
	}
}

// NewAppExitEventListener returns a new app exit event listener.
func NewAppExitEventListener(listener func(me AppExitEvent)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(AppExitEvent); isTyped {
			listener(typed)
		}
	}
}

// AppExitEvent is an event.
type AppExitEvent struct {
	ts  time.Time
	app *App
	err error
}

// Flag returns the logger flag.
func (ae AppExitEvent) Flag() logger.Flag {
	return FlagAppStart
}

// Timestamp returns the timestamp for a
func (ae AppExitEvent) Timestamp() time.Time {
	return ae.ts
}

// App returns the app reference.
func (ae AppExitEvent) App() *App {
	return ae.app
}

// Err retruns an error.
func (ae AppExitEvent) Err() error {
	return ae.err
}

// String implements fmt.Stringer
func (ae AppExitEvent) String() string {
	return "exited"
}

// WriteText implements logger.TextWritable.
func (ae AppExitEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(ae.String())
}

// WriteJSON implements logger.JSONWritable.
func (ae AppExitEvent) WriteJSON() logger.JSONObj {
	return logger.JSONObj{
		logger.JSONFieldMessage: ae.String(),
	}
}
