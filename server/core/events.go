package core

import logger "github.com/blendlabs/go-logger"

var (
	// EventFlagSearch denotes an event.
	EventFlagSearch logger.EventFlag = "giffy.search"

	// EventFlagModeration denotes an event.
	EventFlagModeration logger.EventFlag = "giffy.moderation"

	// EventFlagVote denotes an event.
	EventFlagVote logger.EventFlag = "giffy.vote"
)
