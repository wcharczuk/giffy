package logger

import (
	"sync"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestTimedEventListener(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(2)

	all := New().WithFlags(AllFlags())
	defer all.Close()
	all.Listen(Flag("test-flag"), "default", NewTimedEventListener(func(te *TimedEvent) {
		defer wg.Done()
		assert.Equal("test-flag", te.Flag())
		assert.NotZero(te.Elapsed())
		assert.Equal("foo bar", te.Message())
	}))

	go func() { all.Trigger(Timedf(Flag("test-flag"), time.Millisecond, "foo %s", "bar")) }()
	go func() { all.Trigger(Timedf(Flag("test-flag"), time.Millisecond, "foo %s", "bar")) }()
	wg.Wait()
}
