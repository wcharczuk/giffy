package logger

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// FormatFileSize returns a string representation of a file size in bytes.
func FormatFileSize(sizeBytes int) string {
	if sizeBytes >= 1<<30 {
		return strconv.Itoa(sizeBytes/(1<<30)) + "gB"
	} else if sizeBytes >= 1<<20 {
		return strconv.Itoa(sizeBytes/(1<<20)) + "mB"
	} else if sizeBytes >= 1<<10 {
		return strconv.Itoa(sizeBytes/(1<<10)) + "kB"
	}
	return strconv.Itoa(sizeBytes) + "B"
}

// GetIP gets the origin/client ip for a request.
// X-FORWARDED-FOR is checked. If multiple IPs are included the first one is returned
// X-REAL-IP is checked. If multiple IPs are included the first one is returned
// Finally r.RemoteAddr is used
// Only benevolent services will allow access to the real IP.
func GetIP(r *http.Request) string {
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-FOR", "X-REAL-IP"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

var errTypeConversion = errors.New("Invalid event state type conversion")

func stateAsRequest(state interface{}) (*http.Request, error) {
	if typed, isTyped := state.(*http.Request); isTyped {
		return typed, nil
	}
	return nil, errTypeConversion
}

func stateAsInteger(state interface{}) (int, error) {
	if typed, isTyped := state.(int); isTyped {
		return typed, nil
	}
	return 0, errTypeConversion
}

func stateAsAnsiColorCode(state interface{}) (AnsiColorCode, error) {
	if typed, isTyped := state.(AnsiColorCode); isTyped {
		return typed, nil
	}
	return ColorReset, errTypeConversion
}

func stateAsEventFlag(state interface{}) (uint64, error) {
	if typed, isTyped := state.(uint64); isTyped {
		return typed, nil
	}
	return 0, errTypeConversion
}

func stateAsTime(state interface{}) (time.Time, error) {
	if typed, isTyped := state.(time.Time); isTyped {
		return typed, nil
	}
	return time.Time{}, errTypeConversion
}

func stateAsTimeSource(state interface{}) (TimeSource, error) {
	if typed, isTyped := state.(TimeSource); isTyped {
		return typed, nil
	}
	return SystemClock, errTypeConversion
}

func stateAsDuration(state interface{}) (time.Duration, error) {
	if typed, isTyped := state.(time.Duration); isTyped {
		return typed, nil
	}
	return 0, errTypeConversion
}

func stateAsString(state interface{}) (string, error) {
	if typed, isTyped := state.(string); isTyped {
		return typed, nil
	}
	return "", errTypeConversion
}

func stateAsBytes(state interface{}) ([]byte, error) {
	if typed, isTyped := state.([]byte); isTyped {
		return typed, nil
	}
	return nil, errTypeConversion
}

func envFlagSet(flagName string, defaultValue bool) bool {
	flagValue := os.Getenv(flagName)
	if len(flagValue) > 0 {
		if strings.ToUpper(flagValue) == "TRUE" || flagValue == "1" {
			return true
		}
		return false
	}
	return defaultValue
}
