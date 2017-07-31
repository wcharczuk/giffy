package logger

const (
	// EventAll is a special flag that allows all events to fire.
	EventAll EventFlag = "all"
	// EventNone is a special flag that allows no events to fire.
	EventNone EventFlag = "none"

	// EventFatalError fires for fatal errors (panics or errors returned to users).
	EventFatalError EventFlag = "fatal"
	// EventError fires for errors that are severe enough to log but not so severe as to abort a process.
	EventError EventFlag = "error"
	// EventWarning fires for warnings.
	EventWarning EventFlag = "warning"
	// EventDebug fires for debug messages.
	EventDebug EventFlag = "debug"
	// EventInfo fires for informational messages (app startup etc.)
	EventInfo EventFlag = "info"
	// EventSilly is for when you just need to log something weird.
	EventSilly EventFlag = "silly"

	// EventWebRequestStart fires when an app starts handling a request.
	EventWebRequestStart EventFlag = "web.request.start"
	// EventWebRequest fires when an app completes handling a request.
	EventWebRequest EventFlag = "web.request"
	// EventWebRequestPostBody fires when a request has a post body.
	EventWebRequestPostBody EventFlag = "web.request.postbody"
	// EventWebResponse fires to provide the raw response to a request.
	EventWebResponse EventFlag = "web.response"
)

// EventFlag is a flag to enable or disable triggering handlers for an event.
type EventFlag string
