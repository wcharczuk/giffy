package logger

import (
	"os"
	"strings"
)

// EventFlags
const (
	// EventNone is effectively logging disabled.
	EventNone = uint64(0)
	// EventAll represents every flag being enabled.
	EventAll = ^EventNone
	// EventFatalError enables logging errors
	EventFatalError uint64 = 1 << iota
	// EventError enables logging errors
	EventError uint64 = 1 << iota
	// EventWarning enables logging for warning messages.
	EventWarning uint64 = 1 << iota
	// EventDebug enables logging for debug messages.
	EventDebug uint64 = 1 << iota
	// EventInfo enables logging for informational messages.
	EventInfo uint64 = 1 << iota

	// EventRequest is a helper event for logging request events.
	EventRequest uint64 = 1 << iota
	// EventRequestComplete is a helper event for logging request events with stats.
	EventRequestComplete uint64 = 1 << iota
	// EventRequestBody is a helper event for logging incoming post bodies.
	EventRequestBody uint64 = 1 << iota

	// EventResponse is a helper event for logging response bodies.
	EventResponse uint64 = 1 << iota

	// EventUserError enables output for user error events.
	EventUserError uint64 = 1 << iota
)

// EnvironmentVariables
const (
	// EnvironmentVariableLogVerbosity is the log verbosity environment variable.
	EnvironmentVariableLogVerbosity = "LOG_VERBOSITY"
)

// EventFlagName Lookup
var (
	// EventFlagNames is a map of event flag values to their plaintext names.
	EventFlagNames = map[string]uint64{
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
func ParseEventFlagNameSet(flagValue string) uint64 {
	if len(flagValue) == 0 {
		return EventNone
	}

	flagValueCleaned := strings.Trim(strings.ToUpper(flagValue), " \t\n")
	switch flagValueCleaned {
	case "ALL":
		return EventAll
	case "NONE":
		return EventNone
	}

	return ParseEventNames(strings.Split(flagValue, ",")...)
}

// ParseEventNames parses an array of names into a bit-mask.
func ParseEventNames(flagValues ...string) uint64 {
	result := EventNone
	for _, flagValue := range flagValues {
		result = EventFlagCombine(result, ParseEventName(flagValue))
	}
	return result
}

// ParseEventName parses a single verbosity flag name
func ParseEventName(flagValue string) uint64 {
	flagValueCleaned := strings.Trim(strings.ToUpper(flagValue), " \t\n")
	switch flagValueCleaned {
	case "ALL":
		return EventAll
	case "NONE":
		return EventNone
	default:
		if eventFlag, hasEventFlag := EventFlagNames[flagValueCleaned]; hasEventFlag {
			return eventFlag
		}
		return EventNone
	}
}

// ExpandEventNames expands an event flag set into plaintext names.
func ExpandEventNames(eventFlag uint64) string {
	if eventFlag == EventAll {
		return "ALL"
	}
	if eventFlag == EventNone {
		return "NONE"
	}
	var names []string
	for name, flag := range EventFlagNames {
		if EventFlagAny(eventFlag, flag) {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}

// EventsFromEnvironment parses the environment variable for log verbosity.
func EventsFromEnvironment(defaultEvents ...uint64) uint64 {
	envEventFlag := os.Getenv(EnvironmentVariableLogVerbosity)
	if len(envEventFlag) > 0 {
		return ParseEventFlagNameSet(envEventFlag)
	}
	return EventFlagCombine(defaultEvents...)
}
