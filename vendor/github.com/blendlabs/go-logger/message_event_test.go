package logger

import (
	"sync"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestMessageEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()
	all.Listen(Flag("test-flag"), "default", NewMessageEventListener(func(e *MessageEvent) {
		defer wg.Done()
		assert.Equal("test-flag", e.Flag())
		assert.Equal("foo bar", e.Message())
	}))

	go func() { all.Trigger(Messagef(Flag("test-flag"), "foo %s", "bar")) }()
	go func() { all.Trigger(Messagef(Flag("test-flag"), "foo %s", "bar")) }()
	wg.Wait()
}

func TestMessageEventInterfaces(t *testing.T) {
	assert := assert.New(t)

	ee := Messagef(Info, "this is a test").
		WithHeading("heading").
		WithLabel("foo", "bar")

	eventProvider, isEvent := marshalEvent(ee)
	assert.True(isEvent)
	assert.Equal(Info, eventProvider.Flag())
	assert.False(eventProvider.Timestamp().IsZero())

	headingProvider, isHeadingProvider := marshalEventHeading(ee)
	assert.True(isHeadingProvider)
	assert.Equal("heading", headingProvider.Heading())

	metaProvider, isMetaProvider := marshalEventMeta(ee)
	assert.True(isMetaProvider)
	assert.Equal("bar", metaProvider.Labels()["foo"])
}
