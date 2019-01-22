package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/blend/go-sdk/exception"
)

// Decrypt decrypts data with the given key.
func Decrypt(key, cipherText []byte) ([]byte, error) {
	if len(cipherText) < aes.BlockSize {
		return nil, exception.New("cannot decrypt string: `cipherText` is smaller than AES block size").WithMessagef("block size: %v", aes.BlockSize)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	return cipherText, nil
}
