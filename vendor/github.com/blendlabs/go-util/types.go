package util

import (
	"strings"

	"github.com/blendlabs/go-exception"
)

var (
	// BooleanTrue represents a true value for a boolean.
	BooleanTrue Boolean = true

	// BooleanFalse represents a false value for a boolean.
	BooleanFalse Boolean = false
)

// KeyValuePair is a pair of key and value.
type KeyValuePair struct {
	Key   string
	Value interface{}
}

// KVP is a pair of key and value.
type KVP struct {
	K string
	V interface{}
}

// KeyValuePairOfInt is a pair of key and value.
type KeyValuePairOfInt struct {
	Key   string
	Value int
}

// KVPI is a pair of key and value.
type KVPI struct {
	K string
	V int
}

// KeyValuePairOfFloat is a pair of key and value.
type KeyValuePairOfFloat struct {
	Key   string
	Value float64
}

// KVPF is a pair of key and value.
type KVPF struct {
	K string
	V float64
}

// KeyValuePairOfString is a pair of key and value.
type KeyValuePairOfString struct {
	Key   string
	Value string
}

// KVPS is a pair of key and value.
type KVPS struct {
	K string
	V string
}

// Boolean is a type alias for bool that can be unmarshaled from 0|1, true|false etc.
type Boolean bool

// UnmarshalJSON unmarshals the boolean from json.
func (bit *Boolean) UnmarshalJSON(data []byte) error {
	asString := strings.ToLower(string(data))
	if asString == "1" || asString == "true" {
		*bit = true
		return nil
	} else if asString == "0" || asString == "false" {
		*bit = false
		return nil
	} else if len(asString) > 0 && (asString[0] == '"' || asString[0] == '\'') {
		cleaned := StripQuotes(asString)
		return bit.UnmarshalJSON([]byte(cleaned))
	}
	return exception.Newf("Boolean unmarshal error: invalid input %s", asString)
}

// AsBool returns the stdlib bool value for the boolean.
func (bit Boolean) AsBool() bool {
	if bit {
		return true
	}
	return false
}
