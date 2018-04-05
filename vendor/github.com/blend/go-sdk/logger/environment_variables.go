package logger

//env var names
const (
	// EnvVarLogEvents is the event flag set environment variable.
	EnvVarEventFlags = "LOG_EVENTS"

	// EnvVarLogHiddenEvents is the set of flags that should never produce automatic output.
	EnvVarHiddenEventFlags = "LOG_HIDDEN"

	// EnvVarEvents is the env var that sets the output format.
	EnvVarFormat = "LOG_FORMAT"

	// EnvVarUseColor is the env var that controls if we use ansi colors in output.
	EnvVarUseColor = "LOG_USE_COLOR"
	// EnvVarShowTimestamp is the env var that controls if we show timestamps in output.
	EnvVarShowTime = "LOG_SHOW_TIME"
	// EnvVarShowLabel is the env var that controls if we show a descriptive label in output.
	EnvVarShowLabel = "LOG_SHOW_LABEL"

	// EnvVarLabel is the env var that sets the descriptive label in output.
	EnvVarLabel = "LOG_LABEL"

	// EnvVarTimeFormat is the env var that sets the time format for text output.
	EnvVarTimeFormat = "LOG_TIME_FORMAT"

	// EnvVarJSONPretty returns if we should indent json output.
	EnvVarJSONPretty = "LOG_JSON_PRETTY"
)
