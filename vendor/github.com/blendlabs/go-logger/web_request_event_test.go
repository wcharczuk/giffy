package logger

import (
	"net/http"
	"sync"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestWebRequestEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()
	all.Listen(WebRequest, "default", NewWebRequestEventListener(func(wre *WebRequestEvent) {
		defer wg.Done()
		assert.Equal(WebRequest, wre.Flag())
		assert.NotZero(wre.Elapsed())
		assert.NotNil(wre.Request())
		assert.Equal("test.com", wre.Request().Host)
	}))

	go func() { all.Trigger(NewWebRequestEvent(&http.Request{Host: "test.com"}).WithElapsed(time.Millisecond)) }()
	go func() { all.Trigger(NewWebRequestEvent(&http.Request{Host: "test.com"}).WithElapsed(time.Millisecond)) }()
	wg.Wait()
}

func TestWebRequestEventInterfaces(t *testing.T) {
	assert := assert.New(t)

	ee := NewWebRequestEvent(&http.Request{Host: "test.com"}).WithElapsed(time.Millisecond).WithHeading("heading").WithLabel("foo", "bar")

	eventProvider, isEvent := marshalEvent(ee)
	assert.True(isEvent)
	assert.Equal(WebRequest, eventProvider.Flag())
	assert.False(eventProvider.Timestamp().IsZero())

	headingProvider, isHeadingProvider := marshalEventHeading(ee)
	assert.True(isHeadingProvider)
	assert.Equal("heading", headingProvider.Heading())

	metaProvider, isMetaProvider := marshalEventMeta(ee)
	assert.True(isMetaProvider)
	assert.Equal("bar", metaProvider.Labels()["foo"])
}
