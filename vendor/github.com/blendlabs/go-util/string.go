package util

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	// StringEmpty is the empty string
	StringEmpty = ""

	// ColorRed is the posix escape code fragment for red.
	ColorRed = "31"

	// ColorBlue is the posix escape code fragment for blue.
	ColorBlue = "94"

	// ColorGreen is the posix escape code fragment for green.
	ColorGreen = "32"

	// ColorYellow is the posix escape code fragment for yellow.
	ColorYellow = "33"

	// ColorWhite is the posix escape code fragment for white.
	ColorWhite = "37"

	// ColorGray is the posix escape code fragment for white.
	ColorGray = "90"
)

var (
	// LowerA is the ascii int value for 'a'
	LowerA = uint('a')
	// LowerZ is the ascii int value for 'z'
	LowerZ = uint('z')

	letters           = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numbers           = []rune("0123456789")
	lettersAndNumbers = append(letters, numbers...)
	lowerDiff         = (LowerZ - LowerA)
)

// IsEmpty returns if a string is empty.
func IsEmpty(input string) bool {
	return len(input) == 0
}

// EmptyCoalesce returns the first non-empty string.
func EmptyCoalesce(inputs ...string) string {
	for _, input := range inputs {
		if !IsEmpty(input) {
			return input
		}
	}
	return StringEmpty
}

// CaseInsensitiveEquals compares two strings regardless of case.
func CaseInsensitiveEquals(a, b string) bool {
	aLen := len(a)
	bLen := len(b)
	if aLen != bLen {
		return false
	}

	for x := 0; x < aLen; x++ {
		charA := uint(a[x])
		charB := uint(b[x])

		if charA-LowerA <= lowerDiff {
			charA = charA - 0x20
		}
		if charB-LowerA <= lowerDiff {
			charB = charB - 0x20
		}
		if charA != charB {
			return false
		}
	}

	return true
}

// IsLetter returns if a byte is in the ascii letter range.
func IsLetter(c byte) bool {
	return IsUpper(c) || IsLower(c)
}

// IsUpper returns if a letter is in the ascii upper letter range.
func IsUpper(c byte) bool {
	return c > byte('A') && c < byte('Z')
}

// IsLower returns if a letter is in the ascii lower letter range.
func IsLower(c byte) bool {
	return c > byte('a') && c < byte('z')
}

// CombinePathComponents combines string components of a path.
func CombinePathComponents(components ...string) string {
	slash := "/"
	fullPath := ""
	for index, component := range components {
		workingComponent := component
		if strings.HasPrefix(workingComponent, slash) {
			workingComponent = strings.TrimPrefix(workingComponent, slash)
		}

		if strings.HasSuffix(workingComponent, slash) {
			workingComponent = strings.TrimSuffix(workingComponent, slash)
		}

		if index != len(components)-1 {
			fullPath = fullPath + workingComponent + slash
		} else {
			fullPath = fullPath + workingComponent
		}
	}
	return fullPath
}

// RandomString returns a new random string composed of letters from the `letters` collection.
func RandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// RandomStringWithNumbers returns a random string composed of chars from the `lettersAndNumbers` collection.
func RandomStringWithNumbers(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = lettersAndNumbers[r.Intn(len(lettersAndNumbers))]
	}
	return string(b)
}

// RandomNumbers returns a random string of chars from the `numbers` collection.
func RandomNumbers(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = numbers[r.Intn(len(numbers))]
	}
	return string(b)
}

// IsValidInteger returns if a string is an integer.
func IsValidInteger(input string) bool {
	_, convCrr := strconv.Atoi(input)
	return convCrr == nil
}

// RegexMatch returns if a string matches a regexp.
func RegexMatch(input string, exp string) string {
	regexp := regexp.MustCompile(exp)
	matches := regexp.FindStringSubmatch(input)
	if len(matches) != 2 {
		return StringEmpty
	}
	return strings.TrimSpace(matches[1])
}

// ParseFloat64 parses a float64
func ParseFloat64(input string) float64 {
	result, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0.0
	}
	return result
}

// ParseFloat32 parses a float32
func ParseFloat32(input string) float32 {
	result, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return 0.0
	}
	return float32(result)
}

// ParseInt parses an int
func ParseInt(input string) int {
	result, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}
	return result
}

// ParseInt64 parses an int64
func ParseInt64(input string) int64 {
	result, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return int64(0)
	}
	return result
}

// IntToString turns an int into a string
func IntToString(input int) string {
	return strconv.Itoa(input)
}

// Int64ToString turns an int64 into a string
func Int64ToString(input int64) string {
	return fmt.Sprintf("%v", input)
}

// Float32ToString turns an float32 into a string
func Float32ToString(input float32) string {
	return fmt.Sprintf("%v", input)
}

// Float64ToString turns an float64 into a string
func Float64ToString(input float64) string {
	return fmt.Sprintf("%v", input)
}

// ToCSVOfInt returns a csv from a given slice of integers.
func ToCSVOfInt(input []int) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, IntToString(v))
	}
	return strings.Join(outputStrings, ",")
}

// StripQuotes removes quote characters from a string.
func StripQuotes(input string) string {
	output := []rune{}
	for _, c := range input {
		if !(c == '\'' || c == '"') {
			output = append(output, c)
		}
	}
	return string(output)
}

// TrimWhitespace trims spaces and tabs from a string.
func TrimWhitespace(input string) string {
	return strings.Trim(input, " \t")
}

// IsCamelCase returns if a string is CamelCased.
// CamelCased in this sense is if a string has both upper and lower characters.
func IsCamelCase(input string) bool {
	hasLowers := false
	hasUppers := false

	for _, c := range input {
		if unicode.IsUpper(c) {
			hasUppers = true
		}
		if unicode.IsLower(c) {
			hasLowers = true
		}
	}

	return hasLowers && hasUppers
}

// Base64Encode returns a base64 string for a byte array.
func Base64Encode(blob []byte) string {
	return base64.StdEncoding.EncodeToString(blob)
}

// Base64Decode returns a byte array for a base64 encoded string.
func Base64Decode(blob string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(blob)
}

// Color returns a posix color code escaled string.
func Color(input string, colorCode string) string {
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, input)
}

// ColorFixedWidth returns a posix color code escaled string of a fixed width.
func ColorFixedWidth(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%%d.%ds", width, width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, fixedMessage)
}

// ColorFixedWidthLeftAligned returns a posix color code escaled string of a fixed width left aligned.
func ColorFixedWidthLeftAligned(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, fixedMessage)
}
