package logger

import (
	"strconv"
	"strings"

	exception "github.com/blendlabs/go-exception"
)

const (
	// EventNone is effectively logging disabled.
	EventNone = uint64(0)
	// EventAll represents every flag being enabled.
	EventAll = ^EventNone
	// EventFatalError enables logging errors
	EventFatalError = 1 << iota
	// EventError enables logging errors
	EventError = 1 << iota
	// EventWarning enables logging for warning messages.
	EventWarning = 1 << iota
	// EventDebug enables logging for debug messages.
	EventDebug = 1 << iota
	// EventInfo enables logging for informational messages.
	EventInfo = 1 << iota

	// EventRequest is a helper event for logging request events.
	EventRequest = 1 << iota
	// EventRequestComplete is a helper event for logging request events with stats.
	EventRequestComplete = 1 << iota
	// EventRequestBody is a helper event for logging incoming post bodies.
	EventRequestBody = 1 << iota

	// EventResponse is a helper event for logging response bodies.
	EventResponse = 1 << iota

	// EventUserError enables output for user error events.
	EventUserError = 1 << iota
)

var (
	// EventFlagNames is a map of event flag values to their plaintext names.
	EventFlagNames = map[string]uint64{
		"NONE":                   EventNone,
		"ALL":                    EventAll,
		"LOG_SHOW_FATAL":         EventFatalError,
		"LOG_SHOW_ERROR":         EventError,
		"LOG_SHOW_WARNING":       EventWarning,
		"LOG_SHOW_DEBUG":         EventDebug,
		"LOG_SHOW_INFO":          EventInfo,
		"LOG_SHOW_REQUEST_START": EventRequest,
		"LOG_SHOW_REQUEST":       EventRequestComplete,
		"LOG_SHOW_REQUEST_BODY":  EventRequestBody,
		"LOG_SHOW_RESPONSE":      EventResponse,
		"LOG_SHOW_USER_ERROR":    EventUserError,
	}
)

// EventFlagAll returns if all the reference bits are set for a given value
func EventFlagAll(reference, value uint64) bool {
	return reference&value == value
}

// EventFlagAny returns if any the reference bits are set for a given value
func EventFlagAny(reference, value uint64) bool {
	return reference&value > 0
}

// EventFlagCombine combines all the values into one flag.
func EventFlagCombine(values ...uint64) uint64 {
	var outputFlag uint64
	for _, value := range values {
		outputFlag = outputFlag | value
	}
	return outputFlag
}

// ParseEventFlagNameSet parses an event name csv.
func ParseEventFlagNameSet(flagValue string) (uint64, error) {
	if len(flagValue) == 0 {
		return EventNone, exception.New("Empty `flagValue`")
	}

	if value, parseError := strconv.ParseInt(flagValue, 10, 64); parseError == nil {
		return uint64(value), nil
	}

	return ParseEventNames(strings.Split(flagValue, ",")...)
}

// ParseEventNames parses an array of names into a bit-mask.
func ParseEventNames(flagValues ...string) (uint64, error) {
	var result uint64
	for _, flagValue := range flagValues {
		if parsedValue, parseError := ParseEventName(flagValue); parseError == nil {
			result = result | parsedValue
		} else {
			return result, parseError
		}
	}
	return result, nil
}

// ParseEventName parses a single verbosity flag name
func ParseEventName(flagValue string) (uint64, error) {
	flagValueCleaned := strings.Trim(strings.ToUpper(flagValue), " \t\n")
	switch flagValueCleaned {
	case "ALL":
		return EventAll, nil
	case "NONE":
		return EventNone, nil
	default:
		if eventFlag, hasEventFlag := EventFlagNames[flagValueCleaned]; hasEventFlag {
			return eventFlag, nil
		}
		return EventNone, exception.Newf("Invalid Flag Value: %s", flagValueCleaned)
	}
}

// ExpandEventNames expands an event flag set into plaintext names.
func ExpandEventNames(eventFlag uint64) string {
	if eventFlag == EventNone {
		return "NONE"
	}
	if eventFlag == EventAll {
		return "ALL"
	}
	var names []string
	for name, flag := range EventFlagNames {
		if EventFlagAny(eventFlag, flag) {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}
