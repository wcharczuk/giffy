package logger

import (
	"os"
	"strings"
)

// NewEventFlagSet returns a new EventFlagSet with the given events enabled.
func NewEventFlagSet(eventFlags ...EventFlag) *EventFlagSet {
	efs := &EventFlagSet{
		flags: make(map[EventFlag]bool),
	}
	for _, flag := range eventFlags {
		efs.Enable(flag)
	}
	return efs
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

// NewEventFlagSetFromEnvironment returns a new EventFlagSet from the environment.
func NewEventFlagSetFromEnvironment() *EventFlagSet {
	envEventsFlag := os.Getenv(EnvironmentVariableLogEvents)
	if len(envEventsFlag) > 0 {
		return NewEventFlagSetFromCSV(envEventsFlag)
	}
	return NewEventFlagSet()
}

// NewEventFlagSetFromCSV returns a new event flag set from a csv of event flags.
// These flags are case insensitive.
func NewEventFlagSetFromCSV(flagCSV string) *EventFlagSet {
	flagSet := &EventFlagSet{
		flags: map[EventFlag]bool{},
	}

	flags := strings.Split(flagCSV, ",")

	for _, flag := range flags {
		parsedFlag := EventFlag(strings.Trim(strings.ToLower(flag), " \t\n"))
		if string(parsedFlag) == string(EventAll) {
			flagSet.all = true
		}

		if string(parsedFlag) == string(EventNone) {
			flagSet.none = true
		}

		if strings.HasPrefix(string(parsedFlag), "-") {
			flag := EventFlag(strings.TrimPrefix(string(parsedFlag), "-"))
			flagSet.flags[flag] = false
		} else {
			flagSet.flags[parsedFlag] = true
		}
	}

	return flagSet
}

// EventFlagSet is a set of event flags.
type EventFlagSet struct {
	flags map[EventFlag]bool
	all   bool
	none  bool
}

// Enable enables an event flag.
func (efs *EventFlagSet) Enable(flagValue EventFlag) {
	efs.none = false
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
		// figure out if we explicitly disabled the flag.
		if enabled, hasFlag := efs.flags[flagValue]; hasFlag && !enabled {
			return false
		}
		return true
	}
	if efs.none {
		return false
	}
	if efs.flags != nil {
		if enabled, hasFlag := efs.flags[flagValue]; hasFlag {
			return enabled
		}
	}
	return false
}

func (efs EventFlagSet) String() string {
	if efs.none {
		return string(EventNone)
	}

	var flags []string
	if efs.all {
		flags = []string{string(EventAll)}
	}
	for key, enabled := range efs.flags {
		if key != EventAll {
			if enabled {
				flags = append(flags, string(key))
			} else {
				flags = append(flags, "-"+string(key))
			}
		}
	}
	return strings.Join(flags, ", ")
}
