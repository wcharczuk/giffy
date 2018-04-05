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

// OptionalUInt8 Returns a pointer to a value
func OptionalUInt8(value uint8) *uint8 {
	return &value
}

// OptionalUInt16 Returns a pointer to a value
func OptionalUInt16(value uint16) *uint16 {
	return &value
}

// OptionalUInt Returns a pointer to a value
func OptionalUInt(value uint) *uint {
	return &value
}

// OptionalUInt64 Returns a pointer to a value
func OptionalUInt64(value uint64) *uint64 {
	return &value
}

// OptionalInt16 Returns a pointer to a value
func OptionalInt16(value int16) *int16 {
	return &value
}

// OptionalInt Returns a pointer to a value
func OptionalInt(value int) *int {
	return &value
}

// OptionalInt64 Returns a pointer to a value
func OptionalInt64(value int64) *int64 {
	return &value
}

// OptionalFloat32 Returns a pointer to a value
func OptionalFloat32(value float32) *float32 {
	return &value
}

// OptionalFloat64 Returns a pointer to a value
func OptionalFloat64(value float64) *float64 {
	return &value
}

// OptionalString Returns a pointer to a value
func OptionalString(value string) *string {
	return &value
}

// OptionalBool Returns a pointer to a value
func OptionalBool(value bool) *bool {
	return &value
}

// OptionalTime Returns a pointer to a value
func OptionalTime(value time.Time) *time.Time {
	return &value
}

// OptionalDuration Returns a pointer to a value
func OptionalDuration(value time.Duration) *time.Duration {
	return &value
}
