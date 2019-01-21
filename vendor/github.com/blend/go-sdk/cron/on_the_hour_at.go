package cron

import "time"

var (
	_ Schedule = (*OnTheHourAtUTCSchedule)(nil)
)

// EveryHourOnTheHour returns a schedule that fires every 60 minutes on the 00th minute.
func EveryHourOnTheHour() Schedule {
	return OnTheHourAtUTCSchedule{}
}

// EveryHourAtUTC returns a schedule that fires every hour at a given minute.
func EveryHourAtUTC(minute, second int) Schedule {
	return OnTheHourAtUTCSchedule{Minute: minute, Second: second}
}

// OnTheHourAtUTCSchedule is a schedule that fires every hour on the given minute.
type OnTheHourAtUTCSchedule struct {
	Minute int
	Second int
}

// Next implements the chronometer Schedule api.
func (o OnTheHourAtUTCSchedule) Next(after time.Time) time.Time {
	var returnValue time.Time
	now := Now()
	if after.IsZero() {
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(now) {
			returnValue = returnValue.Add(time.Hour)
		}
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(after) {
			returnValue = returnValue.Add(time.Hour)
		}
	}
	return returnValue
}
