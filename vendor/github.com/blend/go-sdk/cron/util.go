package cron

import "time"

// Now returns a new timestamp.
func Now() time.Time {
	return time.Now().UTC()
}

// Since returns the duration since another timestamp.
func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

// Min returns the minimum of two times.
func Min(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

// Max returns the maximum of two times.
func Max(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}
	return t1
}

// FormatTime returns a string for a time.
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
