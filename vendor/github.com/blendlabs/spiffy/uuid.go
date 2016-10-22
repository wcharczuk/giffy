package spiffy

import (
	"crypto/rand"
	"fmt"
)

// UUID represents a unique identifier conforming to the RFC 4122 standard.
// UUIDs are a fixed 128bit (16 byte) binary blob.
type UUID []byte

// ToFullString returns a "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" hex representation of a uuid.
func (uuid UUID) ToFullString() string {
	b := []byte(uuid)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ToShortString returns a hex representation of the uuid.
func (uuid UUID) ToShortString() string {
	b := []byte(uuid)
	return fmt.Sprintf("%x", b[:])
}

// Version returns the version byte of a uuid.
func (uuid UUID) Version() byte {
	return uuid[6] >> 4
}

// UUIDv4 Create a new UUID version 4.
func UUIDv4() UUID {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // set version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // set variant 10
	return uuid
}
