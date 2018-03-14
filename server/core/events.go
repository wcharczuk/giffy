package core

import logger "github.com/blendlabs/go-logger"

var (
	// FlagSearch denotes an event.
	FlagSearch logger.Flag = "giffy.search"

	// FlagModeration denotes an event.
	FlagModeration logger.Flag = "giffy.moderation"

	// FlagVote denotes an event.
	FlagVote logger.Flag = "giffy.vote"
)
