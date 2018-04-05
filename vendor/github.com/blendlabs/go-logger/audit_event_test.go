package logger

import (
	"sync"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestAuditEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()

	all.Listen(Audit, "default", NewAuditEventListener(func(e *AuditEvent) {
		defer wg.Done()

		assert.Equal(Audit, e.Flag())
		assert.Equal("principal", e.Principal())
		assert.Equal("verb", e.Verb())
	}))

	go func() { all.Trigger(NewAuditEvent("principal", "verb", "noun")) }()
	go func() { all.Trigger(NewAuditEvent("principal", "verb", "noun")) }()
	wg.Wait()
}

func TestAuditEventInterfaces(t *testing.T) {
	assert := assert.New(t)

	ae := NewAuditEvent("principal", "verb", "noun").WithHeading("heading").WithLabel("foo", "bar")

	eventProvider, isEvent := marshalEvent(ae)
	assert.True(isEvent)
	assert.Equal(Audit, eventProvider.Flag())
	assert.False(eventProvider.Timestamp().IsZero())

	headingProvider, isHeadingProvider := marshalEventHeading(ae)
	assert.True(isHeadingProvider)
	assert.Equal("heading", headingProvider.Heading())

	metaProvider, isMetaProvider := marshalEventMeta(ae)
	assert.True(isMetaProvider)
	assert.Equal("bar", metaProvider.Labels()["foo"])
}
