package logger

import "time"

// TimeSource is a type that provides a timestamp.
type TimeSource interface {
	UTCNow() time.Time
}

// SystemClock is the an instance of the system clock timing source.
var SystemClock = timeSourceSystemClock{}

// TimingSourceSystemClock is the system clock timing source.
type timeSourceSystemClock struct{}

// UTCNow returns the current time in UTC.
func (t timeSourceSystemClock) UTCNow() time.Time {
	return time.Now().UTC()
}

// Now returns a historical time instance as a timing source.
func Now() TimeSource {
	return TimeInstance(time.Now())
}

// TimeInstance is the system clock timing source.
type TimeInstance time.Time

// UTCNow returns the current time in UTC.
func (t TimeInstance) UTCNow() time.Time {
	return time.Time(t).UTC()
}
