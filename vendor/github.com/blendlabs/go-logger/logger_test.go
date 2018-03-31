package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util/env"
)

func all(output io.Writer) *Logger {
	return New().WithFlags(AllFlags()).WithWriter(NewTextWriter(output))
}

func TestNewLogger(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	log := all(buffer)
	defer log.Close()

	assert.NotNil(log)
	assert.NotNil(log.Flags())
	assert.True(log.Flags().All())
}

func TestNewLoggerFromEnvironment(t *testing.T) {
	assert := assert.New(t)

	env.Env().Set(EnvVarEventFlags, "all")
	defer env.Env().Restore(EnvVarEventFlags)

	env.Env().Set(EnvVarLabel, "Testing Harness")
	defer env.Env().Restore(EnvVarLabel)

	log := NewFromEnv()
	defer log.Close()

	assert.NotNil(log.Flags())
	assert.Equal("Testing Harness", log.Writers()[0].Label())
}

func TestNewLoggerFromEnvCustomFlags(t *testing.T) {
	assert := assert.New(t)

	env.Env().Set(EnvVarEventFlags, "error,info,web.request")
	defer env.Env().Restore(EnvVarEventFlags)

	env.Env().Set(EnvVarLabel, "Testing Harness")
	defer env.Env().Restore(EnvVarLabel)

	log := NewFromEnv()
	defer log.Close()

	assert.True(log.IsEnabled(Flag("web.request")))
	assert.True(log.IsEnabled(Info))
	assert.False(log.IsEnabled(Warning))
	assert.True(log.IsEnabled(Error))
	assert.False(log.IsEnabled(Fatal))
	assert.Equal("Testing Harness", log.Writers()[0].Label())
}

func TestLoggerEnableDisableEvent(t *testing.T) {
	assert := assert.New(t)

	log := New()
	defer log.Close()
	log.WithEnabled("TEST")
	assert.True(log.IsEnabled("TEST"))
	log.WithEnabled("FOO")
	assert.True(log.IsEnabled("FOO"))

	log.WithDisabled("TEST")
	assert.False(log.IsEnabled("TEST"))
	assert.True(log.IsEnabled("FOO"))
}

func TestLoggerWithFlagSet(t *testing.T) {
	assert := assert.New(t)

	log := New().WithFlags(AllFlags())
	defer log.Close()
	log.WithFlags(NewFlagSet(Info))
	assert.True(log.IsEnabled(Info))
	assert.False(log.IsEnabled(Flag("web.request")))
	assert.False(log.IsHidden(Info))
	log.WithHiddenFlags(NewFlagSet(Info))
	assert.True(log.IsEnabled(Info))
	assert.True(log.IsHidden(Info))
}

func TestLoggerListen(t *testing.T) {
	assert := assert.New(t)

	log := New()
	defer log.Close()

	log.Listen(Error, "foo", func(e Event) {})
	assert.True(log.HasListeners(Error))
	assert.False(log.HasListeners(Fatal))
	assert.True(log.HasListener(Error, "foo"))
	assert.False(log.HasListener(Error, "bar"))
}

func TestLoggerTrigger(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	log := all(buffer)
	defer log.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	var flag Flag
	var contents string
	log.Listen(Error, "foo", func(e Event) {
		defer wg.Done()
		flag = e.Flag()
		contents = fmt.Sprintf("%v", e)
	})
	assert.True(log.IsEnabled(Error))
	assert.True(log.HasListeners(Error))
	assert.True(log.HasListener(Error, "foo"))

	log.Trigger(Messagef(Error, "Hello World"))
	wg.Wait()

	assert.Equal(Error, flag)
	assert.Equal("Hello World", contents)
}

func TestLoggerMultiplexTrigger(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	log := all(buffer)
	defer log.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	log.Listen(Error, "foo", func(e Event) {
		defer wg.Done()
		assert.Equal(Error, e.Flag())
		assert.Equal("Hello World", fmt.Sprintf("%v", e))
	})
	log.Listen(Error, "bar", func(e Event) {
		defer wg.Done()
		assert.Equal(Error, e.Flag())
		assert.Equal("Hello World", fmt.Sprintf("%v", e))
	})
	assert.True(log.IsEnabled(Error))
	assert.True(log.HasListeners(Error))
	assert.True(log.HasListener(Error, "foo"))
	assert.True(log.HasListener(Error, "bar"))
	assert.False(log.HasListener(Error, "baz"))

	log.Trigger(Messagef(Error, "Hello World"))
	wg.Wait()
}

func TestLoggerTriggerUnhandled(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	log := all(buffer)
	defer log.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	log.Listen(Error, "foo", func(e Event) {
		assert.FailNow("The Error Handler shouldn't have fired")
	})
	log.Listen(Fatal, "foo", func(e Event) {
		wg.Done()
	})
	assert.True(log.IsEnabled(Fatal))
	assert.True(log.HasListeners(Fatal))

	log.Trigger(Messagef(WebRequest, "Hello World"))
	log.Trigger(Messagef(Fatal, "Hello World"))
	wg.Wait()
}

func TestLoggerTriggerUnflagged(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer([]byte{})
	log := New(Info, Debug).WithWriter(NewTextWriter(buffer))
	defer log.Close()

	log.Listen(Error, "foo", func(e Event) {
		assert.FailNow("The Error Handler shouldn't have fired")
	})
	assert.False(log.IsEnabled(Error))
	assert.True(log.HasListeners(Error))

	log.Trigger(Messagef(Error, "Hello World"))
}

func TestLoggerRemoveListeners(t *testing.T) {
	assert := assert.New(t)

	log := New()
	defer log.Close()
	log.Listen(Error, "foo", func(e Event) {})
	log.Listen(Info, "foo", func(e Event) {})
	log.RemoveListeners(Error)

	assert.False(log.HasListeners(Error))
	assert.True(log.HasListeners(Info))
}

func TestLoggerSillyf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.Sillyf("foo %s", "bar")

	log.Drain()
	assert.Equal("[silly] foo bar\n", buffer.String())
}

func TestLoggerSyncSillyf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncSillyf("foo %s", "bar")
	assert.Equal("[silly] foo bar\n", buffer.String())
}

func TestLoggerInfof(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.Infof("foo %s", "bar")
	log.Drain()
	assert.Equal("[info] foo bar\n", buffer.String())
}

func TestLoggerSyncInfof(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncInfof("foo %s", "bar")
	assert.Equal("[info] foo bar\n", buffer.String())
}

func TestLoggerDebugf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.Debugf("foo %s", "bar")
	log.Drain()
	assert.Equal("[debug] foo bar\n", buffer.String())
}

func TestLoggerSyncDebugf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncDebugf("foo %s", "bar")
	assert.Equal("[debug] foo bar\n", buffer.String())
}

func TestLoggerWarningf(t *testing.T) {
	assert := assert.New(t)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	writer := NewTextWriter(stdout).
		WithErrorOutput(stderr).
		WithShowTimestamp(false).
		WithUseColor(false)

	log := New().WithFlags(AllFlags()).WithWriter(writer)
	defer log.Close()
	log.Warningf("foo %s", "bar")
	log.Drain()
	assert.Empty(stdout.String())
	assert.Equal("[warning] foo bar\n", stderr.String())
}

func TestLoggerSyncWarningf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncWarningf("foo %s", "bar")
	assert.Equal("[warning] foo bar\n", buffer.String())
}

func TestLoggerErrorf(t *testing.T) {
	assert := assert.New(t)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	writer := NewTextWriter(stdout).
		WithErrorOutput(stderr).
		WithShowTimestamp(false).
		WithUseColor(false)

	log := New().WithFlags(AllFlags()).WithWriter(writer)
	defer log.Close()
	log.Errorf("foo %s", "bar")
	log.Drain()
	assert.Empty(stdout.String())
	assert.Equal("[error] foo bar\n", stderr.String())
}

func TestLoggerSyncErrorf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncErrorf("foo %s", "bar")
	assert.Equal("[error] foo bar\n", buffer.String())
}

func TestLoggerFatalf(t *testing.T) {
	assert := assert.New(t)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	writer := NewTextWriter(stdout).
		WithErrorOutput(stderr).
		WithShowTimestamp(false).
		WithUseColor(false)

	log := New().WithFlags(AllFlags()).WithWriter(writer)
	defer log.Close()
	log.Fatalf("foo %s", "bar")
	log.Drain()
	assert.Empty(stdout.String())
	assert.Equal("[fatal] foo bar\n", stderr.String())
}

func TestLoggerSyncFatalf(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()
	log.SyncFatalf("foo %s", "bar")
	assert.Equal("[fatal] foo bar\n", buffer.String())
}

func TestLoggerJSONErrors(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	writer := NewJSONWriterForOutput(buffer)

	log := New().WithFlags(AllFlags()).WithWriter(writer)
	defer log.Close()
	log.Warningf("foo %s", "bar")
	log.Drain()

	var jsonErr struct {
		Err  Any  `json:"err"`
		Flag Flag `json:"flag"`
	}
	err := json.Unmarshal(buffer.Bytes(), &jsonErr)
	assert.Nil(err)
	assert.Equal("foo bar", jsonErr.Err)
	assert.Equal("warning", jsonErr.Flag)

	buffer.Reset()
	log.Error(exception.Newf("bar %s", "foo"))
	log.Drain()
	var jsonEx struct {
		Err  map[string]Any `json:"err"`
		Flag Flag           `json:"flag"`
	}
	assert.NotEmpty(buffer.Bytes())
	err = json.Unmarshal(buffer.Bytes(), &jsonEx)
	assert.Nil(err)
	assert.NotNil(jsonEx.Err)
	assert.Equal("bar foo", jsonEx.Err["Class"])
	assert.Equal("error", jsonEx.Flag)
}

type enabledEvent struct {
	isEnabled  bool
	isError    bool
	isWritable bool
	message    string
}

func (de enabledEvent) String() string {
	return de.message
}

func (de enabledEvent) Flag() Flag {
	return Flag("enabledEvent")
}

func (de enabledEvent) Timestamp() time.Time {
	return time.Now().UTC()
}

func (de enabledEvent) IsEnabled() bool {
	return de.isEnabled
}

func (de enabledEvent) IsWritable() bool {
	return de.isWritable
}

func (de enabledEvent) IsError() bool {
	return de.isError
}

func TestLoggerEventEnabled(t *testing.T) {
	assert := assert.New(t)
	assert.StartTimeout(500 * time.Millisecond)
	defer assert.EndTimeout()

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	output := make(chan enabledEvent)
	defer close(output)

	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		output <- typed
	})

	log.Trigger(enabledEvent{message: "foo", isEnabled: true, isWritable: true, isError: false})

	received := <-output
	log.Drain()

	assert.Equal("foo", received.message)
	assert.Equal("[enabledEvent] foo\n", buffer.String())
}

func TestLoggerEventDisabled(t *testing.T) {
	assert := assert.New(t)
	assert.StartTimeout(500 * time.Millisecond)
	defer assert.EndTimeout()

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	var didRun bool
	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		_, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		didRun = true
	})

	log.Trigger(enabledEvent{message: "foo", isEnabled: false, isWritable: true, isError: false})
	log.Drain()

	assert.False(didRun)
}

func TestLoggerEventEnabledNotWritten(t *testing.T) {
	assert := assert.New(t)
	assert.StartTimeout(500 * time.Millisecond)
	defer assert.EndTimeout()

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	output := make(chan enabledEvent)
	defer close(output)

	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		output <- typed
	})

	log.Trigger(enabledEvent{message: "foo", isEnabled: true, isWritable: false, isError: false})

	received := <-output
	log.Drain()

	assert.Equal("foo", received.message)
	assert.Empty(buffer.String())
}

func TestLoggerEventEnabledError(t *testing.T) {
	assert := assert.New(t)
	assert.StartTimeout(500 * time.Millisecond)
	defer assert.EndTimeout()

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(stdout).WithErrorOutput(stderr).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	output := make(chan enabledEvent)
	defer close(output)

	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		output <- typed
	})

	log.Trigger(enabledEvent{message: "foo", isEnabled: true, isWritable: true, isError: true})

	received := <-output
	log.Drain()

	assert.Equal("foo", received.message)
	assert.Empty(stdout.String())
	assert.Equal("[enabledEvent] foo\n", stderr.String())
}

func TestLoggerSyncEventEnabled(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	var received enabledEvent
	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		received = typed
	})

	log.SyncTrigger(enabledEvent{message: "foo", isEnabled: true, isWritable: true, isError: false})

	assert.Equal("foo", received.message)
	assert.Equal("[enabledEvent] foo\n", buffer.String())
}

func TestLoggerSyncEventDisabled(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	var received enabledEvent
	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		received = typed
	})

	log.SyncTrigger(enabledEvent{message: "foo", isEnabled: false, isWritable: true, isError: false})
	assert.Empty(received.message)
	assert.Empty(buffer.String())
}

func TestLoggerSyncEventEnabledNotWritten(t *testing.T) {
	assert := assert.New(t)

	buffer := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(buffer).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	var received enabledEvent
	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		received = typed
	})

	log.SyncTrigger(enabledEvent{message: "foo", isEnabled: true, isWritable: false, isError: false})

	assert.Equal("foo", received.message)
	assert.Empty(buffer.String())
}

func TestLoggerSyncEventEnabledError(t *testing.T) {
	assert := assert.New(t)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	log := New().WithFlags(AllFlags()).WithWriter(NewTextWriter(stdout).WithErrorOutput(stderr).WithShowTimestamp(false).WithUseColor(false))
	defer log.Close()

	var received enabledEvent
	log.Listen(Flag("enabledEvent"), DefaultListenerName, func(e Event) {
		typed, isTyped := e.(enabledEvent)
		if !isTyped {
			panic("bad event type")
		}

		received = typed
	})

	log.SyncTrigger(enabledEvent{message: "foo", isEnabled: true, isWritable: true, isError: true})

	assert.Equal("foo", received.message)
	assert.Empty(stdout.String())
	assert.Equal("[enabledEvent] foo\n", stderr.String())
}

func TestLoggerCanConfigureMultipleWriters(t *testing.T) {
	assert := assert.New(t)

	out := bytes.NewBuffer(nil)
	out2 := bytes.NewBuffer(nil)

	log := New().WithWriter(NewTextWriter(out)).WithWriter(NewTextWriter(out2)).WithEnabled(Info)

	log.SyncInfof("this is a %s", "test")
	assert.NotEmpty(out.String())
	assert.NotEmpty(out2.String())
}
