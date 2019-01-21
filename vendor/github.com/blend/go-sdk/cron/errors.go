package cron

import "github.com/blend/go-sdk/exception"

const (
	// ErrJobNotLoaded is a common error.
	ErrJobNotLoaded exception.Class = "job not loaded"

	// ErrJobAlreadyLoaded is a common error.
	ErrJobAlreadyLoaded exception.Class = "job already loaded"

	// ErrJobNotFound is a common error.
	ErrJobNotFound exception.Class = "job not found"

	// ErrJobCancelled is a common error.
	ErrJobCancelled exception.Class = "job cancelled"
)

// IsJobNotLoaded returns if the error is a job not loaded error.
func IsJobNotLoaded(err error) bool {
	return exception.Is(err, ErrJobNotLoaded)
}

// IsJobAlreadyLoaded returns if the error is a job already loaded error.
func IsJobAlreadyLoaded(err error) bool {
	return exception.Is(err, ErrJobAlreadyLoaded)
}

// IsJobNotFound returns if the error is a task not found error.
func IsJobNotFound(err error) bool {
	return exception.Is(err, ErrJobNotFound)
}

// IsJobCancelled returns if the error is a task not found error.
func IsJobCancelled(err error) bool {
	return exception.Is(err, ErrJobCancelled)
}
