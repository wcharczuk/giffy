package logger

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestSeconds(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(1, Seconds(time.Second))
	assert.Equal(2, Seconds(2*time.Second))
	assert.Equal(0.5, Seconds(500*time.Millisecond))
}

func TestMilliseconds(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(1000, Milliseconds(time.Second))
	assert.Equal(2000, Milliseconds(2*time.Second))
	assert.Equal(500, Milliseconds(500*time.Millisecond))
	assert.Equal(0.5, Milliseconds(500*time.Microsecond))
}

func TestMicroseconds(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1000, Microseconds(time.Millisecond))
	assert.Equal(2000, Microseconds(2*time.Millisecond))
	assert.Equal(500, Microseconds(500*time.Microsecond))
	assert.Equal(0.5, Microseconds(500*time.Nanosecond))
}

func TestUnixNano(t *testing.T) {
	assert := assert.New(t)

	actualNanos := (6*time.Millisecond + 7*time.Nanosecond) / time.Nanosecond

	unix, nano := UnixNano(time.Date(2010, 01, 02, 03, 04, 05, int(actualNanos), time.UTC))
	assert.Equal(1262401445, unix)
	assert.Equal(actualNanos, nano)

	checkedDate := time.Unix(unix, nano)
	assert.Equal(2010, checkedDate.Year())
}
