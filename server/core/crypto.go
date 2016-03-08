package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/blendlabs/go-exception"
)

// CreateKey creates a new key for use with the Encrypt or Decrypt methods.
func CreateKey(size int) []byte {
	key := make([]byte, size)
	io.ReadFull(rand.Reader, key)
	return key
}

// Encrypt encrypts the given data with the given key.
func Encrypt(key []byte, text string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(text))
	return ciphertext, nil
}

// Decrypt decrypts the given data with the given key.
func Decrypt(key []byte, cipherText []byte) (string, error) {
	if len(cipherText) < aes.BlockSize {
		return "", exception.New(fmt.Sprintf("Cannot decrypt string: `cipherText` is smaller than AES block size (%v)", aes.BlockSize))
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	return string(cipherText), nil
}
