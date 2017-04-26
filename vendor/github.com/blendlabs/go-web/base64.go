package web

import (
	"encoding/base64"
)

// Base64 is a namespace singleton for base64 functions.
var Base64 base64Util

type base64Util struct{}

func (bu base64Util) Encode(corpus []byte) string {
	return base64.StdEncoding.EncodeToString(corpus)
}

func (bu base64Util) Decode(corpus string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(corpus)
}
