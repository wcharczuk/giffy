package web

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"time"
)

var (
	lowerLetters             = []rune("abcdefghijklmnopqrstuvwxyz")
	upperLetters             = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	letters                  = append(lowerLetters, upperLetters...)
	numbers                  = []rune("0123456789")
	lettersAndNumbers        = append(letters, numbers...)
	symbols                  = []rune(`!@#$%^&*()_+-=[]{}\|:;`)
	lettersNumbersAndSymbols = append(lettersAndNumbers, symbols...)
)

var (
	provider = rand.New(rand.NewSource(time.Now().UnixNano()))
)

var (
	// String is a namespace for string utils.
	String = stringUtil{}
)

type stringUtil struct{}

// RandomRunes returns a random selection of runes from the set.
func (su stringUtil) RandomRunes(runeset []rune, length int) string {
	runes := make([]rune, length)
	runSetLen := len(runeset)
	for index := range runes {
		runes[index] = runeset[provider.Intn(runSetLen)]
	}
	return string(runes)
}

// Random returns a random set of characters.
func (su stringUtil) Random(length int) string {
	return su.RandomRunes(letters, length)
}

// RandomStringWithNumbers returns a random string composed of chars from the `lettersAndNumbers` collection.
func (su stringUtil) RandomWithNumbers(length int) string {
	return su.RandomRunes(lettersAndNumbers, length)
}

// RandomWithNumbersAndSymbols returns a random string composed of chars from the `lettersNumbersAndSymbols` collection.
func (su stringUtil) RandomWithNumbersAndSymbols(length int) string {
	return su.RandomRunes(lettersNumbersAndSymbols, length)
}

// GenerateRandomBytes generates a fixed length of random bytes.
func (su stringUtil) GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (su stringUtil) SecureRandom(length int) string {
	b, _ := su.GenerateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(b)
}
