package oauth

import (
	"encoding/base64"
)

// Base64Decode decodes a string as base64.
func Base64Decode(corpus string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(corpus)
}

// MustBase64Decode decodes a string as base64 and panics if there is an error.
func MustBase64Decode(corpus string) []byte {
	contents, err := Base64Decode(corpus)
	if err != nil {
		panic(err)
	}
	return contents
}

// Base64Encode encodes binary as a base64 string.
func Base64Encode(corpus []byte) string {
	return base64.URLEncoding.EncodeToString(corpus)
}
