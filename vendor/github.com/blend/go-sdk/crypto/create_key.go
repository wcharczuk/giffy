package crypto

import (
	cryptorand "crypto/rand"
)

// MustCreateKey creates a key, if an error is returned, it panics.
func MustCreateKey(keySize int) []byte {
	key, err := CreateKey(keySize)
	if err != nil {
		panic(err)
	}
	return key
}

// CreateKey creates a key of a given size by reading that much data off the crypto/rand reader.
func CreateKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := cryptorand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
