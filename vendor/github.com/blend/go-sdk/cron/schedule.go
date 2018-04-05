package cron

import "time"

// NOTE: time.Zero()? what's that?
var (
	// DaysOfWeek are all the time.Weekday in an array for utility purposes.
	DaysOfWeek = []time.Weekday{
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	}

	// WeekDays are the business time.Weekday in an array.
	WeekDays = []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
	}

	// WeekWeekEndDaysDays are the weekend time.Weekday in an array.
	WeekendDays = []time.Weekday{
		time.Sunday,
		time.Saturday,
	}

	// Epoch is unix epoch saved for utility purposes.
	Epoch = time.Unix(0, 0)
)

// NOTE: we have to use shifts here because in their infinite wisdom google didn't make these values powers of two for masking

const (
	// AllDaysMask is a bitmask of all the days of the week.
	AllDaysMask = 1<<uint(time.Sunday) | 1<<uint(time.Monday) | 1<<uint(time.Tuesday) | 1<<uint(time.Wednesday) | 1<<uint(time.Thursday) | 1<<uint(time.Friday) | 1<<uint(time.Saturday)
	// WeekDaysMask is a bitmask of all the weekdays of the week.
	WeekDaysMask = 1<<uint(time.Monday) | 1<<uint(time.Tuesday) | 1<<uint(time.Wednesday) | 1<<uint(time.Thursday) | 1<<uint(time.Friday)
	//WeekendDaysMask is a bitmask of the weekend days of the week.
	WeekendDaysMask = 1<<uint(time.Sunday) | 1<<uint(time.Saturday)
)

// IsWeekDay returns if the day is a monday->friday.
func IsWeekDay(day time.Weekday) bool {
	return !IsWeekendDay(day)
}

// IsWeekendDay returns if the day is a monday->friday.
func IsWeekendDay(day time.Weekday) bool {
	return day == time.Saturday || day == time.Sunday
}

// Schedule is a type that provides a next runtime after a given previous runtime.
type Schedule interface {
	// GetNextRuntime should return the next runtime after a given previous runtime. If `after` is <nil> it should be assumed
	// the job hasn't run yet. If <nil> is returned by the schedule it is inferred that the job should not run again.
	GetNextRunTime(after *time.Time) *time.Time
}

// EverySecond returns a schedule that fires every second.
func EverySecond() Schedule {
	return IntervalSchedule{Every: 1 * time.Second}
}

// EveryMinute returns a schedule that fires every minute.
func EveryMinute() Schedule {
	return IntervalSchedule{Every: 1 * time.Minute}
}

// EveryHour returns a schedule that fire every hour.
func EveryHour() Schedule {
	return IntervalSchedule{Every: 1 * time.Hour}
}

// Every returns a schedule that fires every given interval.
func Every(interval time.Duration) Schedule {
	return IntervalSchedule{Every: interval}
}

// EveryQuarterHour returns a schedule that fires every 15 minutes, on the quarter hours (0, 15, 30, 45)
func EveryQuarterHour() Schedule {
	return OnTheQuarterHour{}
}

// EveryHourOnTheHour returns a schedule that fires every 60 minutes on the 00th minute.
func EveryHourOnTheHour() Schedule {
	return OnTheHour{}
}

// EveryHourAt returns a schedule that fires every hour at a given minute.
func EveryHourAt(minute int) Schedule {
	return OnTheHourAt{minute}
}

// WeeklyAt returns a schedule that fires on every of the given days at the given time by hour, minute and second.
func WeeklyAt(hour, minute, second int, days ...time.Weekday) Schedule {
	dayOfWeekMask := uint(0)
	for _, day := range days {
		dayOfWeekMask = dayOfWeekMask | 1<<uint(day)
	}

	return &DailySchedule{DayOfWeekMask: dayOfWeekMask, TimeOfDayUTC: time.Date(0, 0, 0, hour, minute, second, 0, time.UTC)}
}

// DailyAt returns a schedule that fires every day at the given hour, minut and second.
func DailyAt(hour, minute, second int) Schedule {
	return &DailySchedule{DayOfWeekMask: AllDaysMask, TimeOfDayUTC: time.Date(0, 0, 0, hour, minute, second, 0, time.UTC)}
}

// WeekdaysAt returns a schedule that fires every week day at the given hour, minut and second.
func WeekdaysAt(hour, minute, second int) Schedule {
	return &DailySchedule{DayOfWeekMask: WeekDaysMask, TimeOfDayUTC: time.Date(0, 0, 0, hour, minute, second, 0, time.UTC)}
}

// WeekendsAt returns a schedule that fires every weekend day at the given hour, minut and second.
func WeekendsAt(hour, minute, second int) Schedule {
	return &DailySchedule{DayOfWeekMask: WeekendDaysMask, TimeOfDayUTC: time.Date(0, 0, 0, hour, minute, second, 0, time.UTC)}
}

// --------------------------------------------------------------------------------
// Schedule Implementations
// --------------------------------------------------------------------------------

// OnDemand returns an on demand schedule, or a schedule that only allows the job to be run
// explicitly by calling `RunJob` on the `JobManager`.
func OnDemand() Schedule {
	return OnDemandSchedule{}
}

// OnDemandSchedule is a schedule that runs on demand.
type OnDemandSchedule struct{}

// GetNextRunTime gets the next run time.
func (ods OnDemandSchedule) GetNextRunTime(after *time.Time) *time.Time {
	return nil
}

// Immediately Returns a schedule that casues a job to run immediately on start,
// with an optional subsequent schedule.
func Immediately() *ImmediateSchedule {
	return &ImmediateSchedule{}
}

// ImmediateSchedule fires immediately with an optional subsequent schedule..
type ImmediateSchedule struct {
	didRun bool
	then   Schedule
}

// Then allows you to specify a subsequent schedule after the first run.
func (i *ImmediateSchedule) Then(then Schedule) Schedule {
	i.then = then
	return i
}

// GetNextRunTime implements Schedule.
func (i *ImmediateSchedule) GetNextRunTime(after *time.Time) *time.Time {
	if !i.didRun {
		i.didRun = true
		return optional(Now())
	}
	if i.then != nil {
		return i.then.GetNextRunTime(after)
	}
	return optional(Now())
}

// IntervalSchedule is as chedule that fires every given interval with an optional start delay.
type IntervalSchedule struct {
	Every      time.Duration
	StartDelay *time.Duration
}

// GetNextRunTime implements Schedule.
func (i IntervalSchedule) GetNextRunTime(after *time.Time) *time.Time {
	if after == nil {
		if i.StartDelay == nil {
			next := Now().Add(i.Every)
			return &next
		}
		next := Now().Add(*i.StartDelay).Add(i.Every)
		return &next
	}
	last := *after
	last = last.Add(i.Every)
	return &last
}

// DailySchedule is a schedule that fires every day that satisfies the DayOfWeekMask at the given TimeOfDayUTC.
type DailySchedule struct {
	DayOfWeekMask uint
	TimeOfDayUTC  time.Time
}

func (ds DailySchedule) checkDayOfWeekMask(day time.Weekday) bool {
	trialDayMask := uint(1 << uint(day))
	bitwiseResult := (ds.DayOfWeekMask & trialDayMask)
	return bitwiseResult > uint(0)
}

// GetNextRunTime implements Schedule.
func (ds DailySchedule) GetNextRunTime(after *time.Time) *time.Time {
	if after == nil {
		after = optional(Now())
	}

	todayInstance := time.Date(after.Year(), after.Month(), after.Day(), ds.TimeOfDayUTC.Hour(), ds.TimeOfDayUTC.Minute(), ds.TimeOfDayUTC.Second(), 0, time.UTC)
	for day := 0; day < 8; day++ {
		next := todayInstance.AddDate(0, 0, day) //the first run here it should be adding nothing, i.e. returning todayInstance ...

		if ds.checkDayOfWeekMask(next.Weekday()) && next.After(*after) { //we're on a day ...
			return &next
		}
	}

	return &Epoch
}

// OnTheQuarterHour is a schedule that fires every 15 minutes, on the quarter hours.
type OnTheQuarterHour struct{}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheQuarterHour) GetNextRunTime(after *time.Time) *time.Time {
	var returnValue time.Time
	if after == nil {
		now := Now()
		if now.Minute() >= 45 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 30 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 15 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	} else {
		if after.Minute() >= 45 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 30 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 15 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	}
	return &returnValue
}

// OnTheHour is a schedule that fires every hour on the 00th minute.
type OnTheHour struct{}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheHour) GetNextRunTime(after *time.Time) *time.Time {
	var returnValue time.Time
	now := Now()
	if after == nil {
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
		if returnValue.Before(now) {
			returnValue = returnValue.Add(time.Hour)
		}
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
	}
	return &returnValue
}

// OnTheHourAt is a schedule that fires every hour on the given minute.
type OnTheHourAt struct {
	Minute int
}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheHourAt) GetNextRunTime(after *time.Time) *time.Time {
	var returnValue time.Time
	now := Now()
	if after == nil {
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), o.Minute, 0, 0, time.UTC)
		if returnValue.Before(now) {
			returnValue = returnValue.Add(time.Hour)
		}
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), o.Minute, 0, 0, time.UTC)
		if returnValue.Before(*after) {
			returnValue = returnValue.Add(time.Hour)
		}
	}
	return &returnValue
}

// --------------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------------

func optional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
