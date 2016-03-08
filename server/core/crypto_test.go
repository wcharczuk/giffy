package core

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestEncryptDecrypt(t *testing.T) {
	a := assert.New(t)

	text := "this is a test"
	key := CreateKey(32)
	cipherText, cipherErr := Encrypt(key, text)
	a.Nil(cipherErr)
	a.NotEmpty(cipherText)
	decrypted, decryptErr := Decrypt(key, cipherText)
	a.Nil(decryptErr)
	a.Equal(text, decrypted)
}
