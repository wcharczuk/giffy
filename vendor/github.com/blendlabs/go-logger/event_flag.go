package logger

import (
	"os"
	"strings"
)

const (
	// EventAll is a special flag that allows all events to fire.
	EventAll EventFlag = "ALL"
	// EventNone is a special flag that allows no events to fire.
	EventNone EventFlag = "NONE"

	// EventFatalError fires for fatal errors (panics or errors returned to users).
	EventFatalError EventFlag = "FATAL"
	// EventError fires for errors that are severe enough to log but not so severe as to abort a process.
	EventError EventFlag = "ERROR"
	// EventWarning fires for warnings.
	EventWarning EventFlag = "WARNING"
	// EventDebug fires for debug messages.
	EventDebug EventFlag = "DEBUG"
	// EventInfo fires for informational messages (app startup etc.)
	EventInfo EventFlag = "INFO"
	// EventRequest fires when an app starts handling a request.
	EventRequest EventFlag = "REQUEST"
	// EventRequestComplete fires when an app completes handling a request.
	EventRequestComplete EventFlag = "REQUEST_COMPLETE"
	// EventRequestPostBody fires when a request has a post body.
	EventRequestPostBody EventFlag = "REQUEST_POST_BODY"
	// EventResponse fires to provide the raw response to a request.
	EventResponse EventFlag = "RESPONSE"
	// EventUserError is a particular class of error caused by callers of a service.
	EventUserError EventFlag = "USER_ERROR"

	// EventInternalError is an alias to EventFatalError
	EventInternalError = EventFatalError
)

// EnvironmentVariables
const (
	// EnvironmentVariableLogEvents is the log verbosity environment variable.
	EnvironmentVariableLogEvents = "LOG_EVENTS"
)

var (
	// AllEventFlags is an array of all the event flags.
	AllEventFlags = []EventFlag{
		EventFatalError,
		EventError,
		EventWarning,
		EventDebug,
		EventInfo,
		EventRequest,
		EventRequestComplete,
		EventResponse,
		EventUserError,
	}
)

// EventFlag is a flag to enable or disable triggering handlers for an event.
type EventFlag string

// NewEventFlagSet returns a new EventFlagSet.
func NewEventFlagSet() *EventFlagSet {
	return &EventFlagSet{
		flags: make(map[EventFlag]bool),
	}
}

// NewEventFlagSetAll returns a new EventFlagSet with all flags enabled.
func NewEventFlagSetAll() *EventFlagSet {
	return &EventFlagSet{
		flags: make(map[EventFlag]bool),
		all:   true,
	}
}

// NewEventFlagSetNone returns a new EventFlagSet with no flags enabled.
func NewEventFlagSetNone() *EventFlagSet {
	return &EventFlagSet{
		flags: make(map[EventFlag]bool),
		none:  true,
	}
}

// NewEventFlagSetWithEvents returns a new EventFlagSet with the given events enabled.
func NewEventFlagSetWithEvents(eventFlags ...EventFlag) *EventFlagSet {
	efs := &EventFlagSet{
		flags: make(map[EventFlag]bool),
	}
	for _, flag := range eventFlags {
		efs.Enable(flag)
	}
	return efs
}

// NewEventFlagSetFromEnvironment returns a new EventFlagSet from the environment.
func NewEventFlagSetFromEnvironment() *EventFlagSet {
	envEventsFlag := os.Getenv(EnvironmentVariableLogEvents)
	if len(envEventsFlag) > 0 {
		flags := strings.Split(envEventsFlag, ",")
		var events []EventFlag
		for _, flag := range flags {
			parsedFlag := EventFlag(strings.Trim(strings.ToUpper(flag), " \t\n"))
			if CaseInsensitiveEquals(string(parsedFlag), string(EventAll)) {
				return NewEventFlagSetAll()
			}
			if CaseInsensitiveEquals(string(parsedFlag), string(EventNone)) {
				return NewEventFlagSetNone()
			}
			events = append(events, parsedFlag)
		}
		return NewEventFlagSetWithEvents(events...)
	}
	return NewEventFlagSet()
}

// EventFlagSet is a set of event flags.
type EventFlagSet struct {
	flags map[EventFlag]bool
	all   bool
	none  bool
}

// Enable enables an event flag.
func (efs *EventFlagSet) Enable(flagValue EventFlag) {
	efs.flags[flagValue] = true
}

// Disable disabled an event flag.
func (efs *EventFlagSet) Disable(flagValue EventFlag) {
	efs.flags[flagValue] = false
}

// EnableAll flips the `all` bit on the flag set.
func (efs *EventFlagSet) EnableAll() {
	efs.all = true
	efs.none = false
}

// IsAllEnabled returns if the all bit is flipped on.
func (efs *EventFlagSet) IsAllEnabled() bool {
	return efs.all
}

// IsNoneEnabled returns if the none bit is flipped on.
func (efs *EventFlagSet) IsNoneEnabled() bool {
	return efs.none
}

// DisableAll flips the `none` bit on the flag set.
func (efs *EventFlagSet) DisableAll() {
	efs.all = false
	efs.none = true
}

// IsEnabled checks to see if an event is enabled.
func (efs EventFlagSet) IsEnabled(flagValue EventFlag) bool {
	if efs.all {
		return true
	}
	if efs.none {
		return false
	}
	if enabled, hasFlag := efs.flags[flagValue]; hasFlag {
		return enabled
	}
	return false
}

func (efs EventFlagSet) String() string {
	if efs.all {
		return string(EventAll)
	}
	if efs.none {
		return string(EventNone)
	}
	var flags []string
	for key, enabled := range efs.flags {
		if enabled {
			flags = append(flags, string(key))
		}
	}
	return strings.Join(flags, ", ")
}
