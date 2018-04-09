package web

import (
	"bytes"
	"fmt"
	"time"

	"github.com/blend/go-sdk/logger"
)

const (
	// AppStart fires when the app is starting.
	AppStart logger.Flag = "web.app.start"
	// AppStartComplete fires after the app has started.
	AppStartComplete logger.Flag = "web.app.start.complete"
	// AppExit fires when an app exits.
	AppExit logger.Flag = "web.app.exit"
)

const (
	// HealthzStart is a logger event.
	HealthzStart logger.Flag = "web.healthz.start"
	// HealthzStartComplete is a logger event.
	HealthzStartComplete logger.Flag = "web.healthz.start.complete"
	// HealthzExit is a logger event.
	HealthzExit logger.Flag = "web.healthz.exit"
)

const (
	// HTTPSUpgraderStart is a logger event.
	HTTPSUpgraderStart logger.Flag = "web.upgrader.start"
	// HTTPSUpgraderStartComplete is a logger event.
	HTTPSUpgraderStartComplete logger.Flag = "web.upgrader.start.complete"
	// HTTPSUpgraderExit is a logger event.
	HTTPSUpgraderExit logger.Flag = "web.upgrader.exit"
)

// NewAppEvent creates a new app start event.
func NewAppEvent(flag logger.Flag) *AppEvent {
	return &AppEvent{
		flag: flag,
		ts:   time.Now().UTC(),
	}
}

// NewAppEventListener returns a new app start event listener.
func NewAppEventListener(listener func(me *AppEvent)) logger.Listener {
	return func(e logger.Event) {
		if typed, isTyped := e.(*AppEvent); isTyped {
			listener(typed)
		}
	}
}

// AppEvent is an event.
type AppEvent struct {
	flag     logger.Flag
	ts       time.Time
	app      *App
	hz       *Healthz
	upgrader *HTTPSUpgrader
	elapsed  time.Duration
	err      error
}

// WithFlag sets the event flag.
func (ae *AppEvent) WithFlag(flag logger.Flag) *AppEvent {
	ae.flag = flag
	return ae
}

// Flag returns the logger flag.
func (ae *AppEvent) Flag() logger.Flag {
	return ae.flag
}

// WithTimestamp sets the event timestamp.
func (ae *AppEvent) WithTimestamp(ts time.Time) *AppEvent {
	ae.ts = ts
	return ae
}

// Timestamp returns the timestamp for a
func (ae *AppEvent) Timestamp() time.Time {
	return ae.ts
}

// WithApp sets the event app reference.
func (ae *AppEvent) WithApp(app *App) *AppEvent {
	ae.app = app
	return ae
}

// App returns the app reference.
func (ae AppEvent) App() *App {
	return ae.app
}

// WithHealthz sets the event hz reference.
func (ae *AppEvent) WithHealthz(hz *Healthz) *AppEvent {
	ae.hz = hz
	return ae
}

// Healthz returns the healthz reference.
func (ae AppEvent) Healthz() *Healthz {
	return ae.hz
}

// WithUpgrader sets the event hz reference.
func (ae *AppEvent) WithUpgrader(upgrader *HTTPSUpgrader) *AppEvent {
	ae.upgrader = upgrader
	return ae
}

// Upgrader returns the https upgrader reference.
func (ae AppEvent) Upgrader() *HTTPSUpgrader {
	return ae.upgrader
}

// WithErr sets the event error.
func (ae *AppEvent) WithErr(err error) *AppEvent {
	ae.err = err
	return ae
}

// Err returns an underlying error.
func (ae *AppEvent) Err() error {
	return ae.err
}

// WithElapsed sets the elapsed time on the event.
func (ae *AppEvent) WithElapsed(elapsed time.Duration) *AppEvent {
	ae.elapsed = elapsed
	return ae
}

// Elapsed returns the elapsed time.
func (ae *AppEvent) Elapsed() time.Duration {
	return ae.elapsed
}

// WriteText implements logger.TextWritable.
func (ae *AppEvent) WriteText(tf logger.TextFormatter, buf *bytes.Buffer) {
	if ae.elapsed > 0 {
		if ae.err != nil {
			buf.WriteString(tf.Colorize("failed", logger.ColorRed))
			buf.WriteRune(logger.RuneNewline)
			buf.WriteString(fmt.Sprintf("%+v", ae.err))
		} else {
			buf.WriteString(tf.Colorize("complete", logger.ColorBlue))
		}

		buf.WriteRune(logger.RuneSpace)
		buf.WriteString(fmt.Sprintf("(%v)", ae.elapsed))
	}
}

// WriteJSON implements logger.JSONWritable.
func (ae *AppEvent) WriteJSON() logger.JSONObj {
	obj := logger.JSONObj{}
	if ae.err != nil {
		obj[logger.JSONFieldErr] = ae.err
	}
	if ae.elapsed > 0 {
		obj[logger.JSONFieldElapsed] = ae.elapsed
	}
	return obj
}
