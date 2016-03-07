package core

import (
	"crypto/rand"
	"fmt"
)

// UUID is a fixed length byte array.
type UUID []byte

// ToFullString returns a long form of a UUID.
func (uuid UUID) ToFullString() string {
	b := []byte(uuid)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ToShortString returns a short form of a UUID.
func (uuid UUID) ToShortString() string {
	b := []byte(uuid)
	return fmt.Sprintf("%x", b[:])
}

// Version returns the uuid version of a uuid.
func (uuid UUID) Version() byte {
	return uuid[6] >> 4
}

// UUIDv4 returns a version 4 uuid.
func UUIDv4() UUID {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid
}
