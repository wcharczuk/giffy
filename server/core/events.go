package core

import logger "github.com/blendlabs/go-logger"

var (
	// EventFlagSearch denotes an event.
	EventFlagSearch logger.EventFlag = "SEARCH"

	// EventFlagModeration denotes an event.
	EventFlagModeration logger.EventFlag = "MODERATION"

	// EventFlagVote denotes an event.
	EventFlagVote logger.EventFlag = "VOTE"
)
