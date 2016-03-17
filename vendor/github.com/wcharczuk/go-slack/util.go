package slack

import "time"

func OptionalUInt8(value uint8) *uint8 {
	return &value
}

func OptionalByte(value byte) *byte {
	return &value
}

func OptionalUInt16(value uint16) *uint16 {
	return &value
}

func OptionalUInt(value uint) *uint {
	return &value
}

func OptionalUInt32(value uint32) *uint32 {
	return &value
}

func OptionalUInt64(value uint64) *uint64 {
	return &value
}

func OptionalInt16(value int16) *int16 {
	return &value
}

func OptionalInt(value int) *int {
	return &value
}

func OptionalInt32(value int32) *int32 {
	return &value
}

func OptionalInt64(value int64) *int64 {
	return &value
}

func OptionalFloat32(value float32) *float32 {
	return &value
}

func OptionalFloat64(value float64) *float64 {
	return &value
}

func OptionalString(value string) *string {
	return &value
}

func OptionalBool(value bool) *bool {
	return &value
}

func OptionalTime(value time.Time) *time.Time {
	return &value
}

func OptionalTimestamp(value Timestamp) *Timestamp {
	return &value
}

func IsEmpty(s string) bool {
	return len(s) == 0
}
