package util

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	// StringEmpty is the empty string
	StringEmpty = ""

	// ColorBlack is the posix escape code fragment for black.
	ColorBlack = "30m"

	// ColorRed is the posix escape code fragment for red.
	ColorRed = "31m"

	// ColorGreen is the posix escape code fragment for green.
	ColorGreen = "32m"

	// ColorYellow is the posix escape code fragment for yellow.
	ColorYellow = "33m"

	// ColorBlue is the posix escape code fragment for blue.
	ColorBlue = "34m"

	// ColorPurple is the posix escape code fragement for magenta (purple)
	ColorPurple = "35m"

	// ColorCyan is the posix escape code fragement for cyan.
	ColorCyan = "36m"

	// ColorWhite is the posix escape code fragment for white.
	ColorWhite = "37m"

	// ColorLightBlack is the posix escape code fragment for black.
	ColorLightBlack = "90m"

	// ColorLightRed is the posix escape code fragment for red.
	ColorLightRed = "91m"

	// ColorLightGreen is the posix escape code fragment for green.
	ColorLightGreen = "92m"

	// ColorLightYellow is the posix escape code fragment for yellow.
	ColorLightYellow = "93m"

	// ColorLightBlue is the posix escape code fragment for blue.
	ColorLightBlue = "94m"

	// ColorLightPurple is the posix escape code fragement for magenta (purple)
	ColorLightPurple = "95m"

	// ColorLightCyan is the posix escape code fragement for cyan.
	ColorLightCyan = "96m"

	// ColorLightWhite is the posix escape code fragment for white.
	ColorLightWhite = "97m"

	// ColorGray is an alias to ColorLightWhite to preserve backwards compatibility.
	ColorGray = ColorLightBlack

	// ColorReset is the posix escape code fragment to reset all formatting.
	ColorReset = "0m"
)

var (
	// LowerA is the ascii int value for 'a'
	LowerA = uint('a')
	// LowerZ is the ascii int value for 'z'
	LowerZ = uint('z')

	lowerDiff = (LowerZ - LowerA)

	// LowerLetters is a runset of lowercase letters.
	LowerLetters = []rune("abcdefghijklmnopqrstuvwxyz")

	// UpperLetters is a runset of uppercase letters.
	UpperLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Letters is a runset of both lower and uppercase letters.
	Letters = append(LowerLetters, UpperLetters...)

	// Numbers is a runset of numeric characters.
	Numbers = []rune("0123456789")

	// LettersAndNumbers is a runset of letters and numeric characters.
	LettersAndNumbers = append(Letters, Numbers...)

	// Symbols is a runset of symbol characters.
	Symbols = []rune(`!@#$%^&*()_+-=[]{}\|:;`)

	// LettersNumbersAndSymbols is a runset of letters, numbers and symbols.
	LettersNumbersAndSymbols = append(LettersAndNumbers, Symbols...)
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

// IsLetter returns if a rune is in the ascii letter range.
func IsLetter(c rune) bool {
	return IsUpper(c) || IsLower(c)
}

// IsUpper returns if a rune is in the ascii upper letter range.
func IsUpper(c rune) bool {
	return c >= rune('A') && c <= rune('Z')
}

// IsLower returns if a rune is in the ascii lower letter range.
func IsLower(c rune) bool {
	return c >= rune('a') && c <= rune('z')
}

// IsNumber returns if a rune is in the number range.
func IsNumber(c rune) bool {
	return c >= rune('0') && c <= rune('9')
}

// IsSymbol returns if the rune is in the symbol range.
func IsSymbol(c rune) bool {
	return c >= rune(' ') && c <= rune('/')
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

// StringAny returns true if any of the possibles are == to the basis.
func StringAny(basis string, possibles ...string) bool {
	for _, possible := range possibles {
		if basis == possible {
			return true
		}
	}
	return false
}

// StringAnyInsensitive returns true if any of the possibles are == to the basis regardless of letter casing.
func StringAnyInsensitive(basis string, possibles ...string) bool {
	for _, possible := range possibles {
		if CaseInsensitiveEquals(basis, possible) {
			return true
		}
	}
	return false
}

// RandomString returns a new random string composed of letters from the `letters` collection.
func RandomString(length int) string {
	return RandomRunes(Letters, length)
}

// RandomNumbers returns a random string of chars from the `numbers` collection.
func RandomNumbers(length int) string {
	return RandomRunes(Numbers, length)
}

// RandomStringWithNumbers returns a random string composed of chars from the `lettersAndNumbers` collection.
func RandomStringWithNumbers(length int) string {
	return RandomRunes(LettersAndNumbers, length)
}

// RandomStringWithNumbersAndSymbols returns a random string composed of chars from the `lettersNumbersAndSymbols` collection.
func RandomStringWithNumbersAndSymbols(length int) string {
	return RandomRunes(LettersNumbersAndSymbols, length)
}

// RandomRunes returns a random selection of runes from the set.
func RandomRunes(runeset []rune, length int) string {
	runes := make([]rune, length)
	for index := range runes {
		runes[index] = runeset[provider.Intn(len(runeset))]
	}
	return string(runes)
}

// CombineRunsets combines given runsets into a single runset.
func CombineRunsets(runesets ...[]rune) []rune {
	output := []rune{}
	for _, set := range runesets {
		output = append(output, set...)
	}
	return output
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

//AnsiEscapeCode prefixes a color or text formatting code with the ESC keyboard code and a `[` character.
func AnsiEscapeCode(code string) string {
	return fmt.Sprintf("\033[%s", code)
}

// Color returns a posix color code escaled string.
func Color(input string, colorCode string) string {
	return fmt.Sprintf("%s%s%s", AnsiEscapeCode(colorCode), input, AnsiEscapeCode(ColorReset))
}

// ColorFixedWidth returns a posix color code escaled string of a fixed width.
func ColorFixedWidth(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%%d.%ds", width, width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("%s%s%s", AnsiEscapeCode(colorCode), fixedMessage, AnsiEscapeCode(ColorReset))
}

// ColorFixedWidthLeftAligned returns a posix color code escaled string of a fixed width left aligned.
func ColorFixedWidthLeftAligned(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("%s%s%s", AnsiEscapeCode(colorCode), fixedMessage, AnsiEscapeCode(ColorReset))
}
