package logger

import (
	"sync"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestQueryEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()

	all.Listen(Query, "default", NewQueryEventListener(func(e *QueryEvent) {
		defer wg.Done()

		assert.Equal(Query, e.Flag())
		assert.Equal("moo", e.QueryLabel())
		assert.Equal("foo bar", e.Body())
	}))

	go func() { all.Trigger(NewQueryEvent("foo bar", time.Second).WithQueryLabel("moo")) }()
	go func() { all.Trigger(NewQueryEvent("foo bar", time.Second).WithQueryLabel("moo")) }()
	wg.Wait()
}

func TestQueryEventInterfaces(t *testing.T) {
	assert := assert.New(t)

	ee := NewQueryEvent("this is a test", time.Second).
		WithHeading("heading").
		WithLabel("foo", "bar")

	eventProvider, isEvent := marshalEvent(ee)
	assert.True(isEvent)
	assert.Equal(Query, eventProvider.Flag())
	assert.False(eventProvider.Timestamp().IsZero())

	headingProvider, isHeadingProvider := marshalEventHeading(ee)
	assert.True(isHeadingProvider)
	assert.Equal("heading", headingProvider.Heading())

	metaProvider, isMetaProvider := marshalEventMeta(ee)
	assert.True(isMetaProvider)
	assert.Equal("bar", metaProvider.Labels()["foo"])
}
