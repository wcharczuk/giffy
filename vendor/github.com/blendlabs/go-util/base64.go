package util

import "encoding/base64"

type base64Util struct{}

// Encode returns a base64 string for a byte array.
func (bu base64Util) Encode(blob []byte) string {
	return base64.StdEncoding.EncodeToString(blob)
}

// Decode returns a byte array for a base64 encoded string.
func (bu base64Util) Decode(blob string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(blob)
}
