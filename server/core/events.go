package core

import logger "github.com/blendlabs/go-logger"

var (
	// EventFlagSearch denotes an event.
	EventFlagSearch = logger.CreateEventFlagConstant(0)

	// EventFlagModeration denotes an event.
	EventFlagModeration = logger.CreateEventFlagConstant(1)

	// EventFlagVote denotes an event.
	EventFlagVote = logger.CreateEventFlagConstant(2)
)
