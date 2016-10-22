package spiffy

import "time"

// DbEvent is a stats struct for db events.
type DbEvent struct {
	DbConnection *DbConnection
	Query        string
	Error        error
	Elapsed      time.Duration
}

// DbEventListener is an event listener for DB events.
type DbEventListener func(dbe *DbEvent)
