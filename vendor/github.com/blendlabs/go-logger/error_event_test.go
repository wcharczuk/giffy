package logger

import (
	"fmt"
	"sync"
	"testing"

	assert "github.com/blendlabs/go-assert"
	exception "github.com/blendlabs/go-exception"
)

func TestErrorEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()

	all.Listen(Fatal, "default", NewErrorEventListener(func(e *ErrorEvent) {
		defer wg.Done()

		assert.Equal(Fatal, e.Flag())
		assert.Equal("foo bar", e.Err().Error())
	}))

	go func() { all.Trigger(NewErrorEvent(Fatal, fmt.Errorf("foo bar"))) }()
	go func() { all.Trigger(NewErrorEvent(Fatal, fmt.Errorf("foo bar"))) }()
	wg.Wait()
}

func TestErrorEventInterfaces(t *testing.T) {
	assert := assert.New(t)

	ee := NewErrorEvent(Fatal,
		exception.New("this is a test").
			WithMessagef("this is a message").
			WithStack(exception.StackStrings([]string{"foo", "bar"}))).
		WithHeading("heading").
		WithLabel("foo", "bar")

	eventProvider, isEvent := marshalEvent(ee)
	assert.True(isEvent)
	assert.Equal(Fatal, eventProvider.Flag())
	assert.False(eventProvider.Timestamp().IsZero())

	headingProvider, isHeadingProvider := marshalEventHeading(ee)
	assert.True(isHeadingProvider)
	assert.Equal("heading", headingProvider.Heading())

	metaProvider, isMetaProvider := marshalEventMeta(ee)
	assert.True(isMetaProvider)
	assert.Equal("bar", metaProvider.Labels()["foo"])
}
