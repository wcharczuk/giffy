package cron

import (
	"sync/atomic"
)

// AtomicFlag is a boolean value that is syncronized.
type AtomicFlag struct {
	value int32
}

// Set the flag value.
func (af *AtomicFlag) Set(value bool) {
	if value {
		atomic.StoreInt32(&af.value, 1)
	} else {
		atomic.StoreInt32(&af.value, 0)
	}
}

// Get the flag value.
func (af *AtomicFlag) Get() (value bool) {
	value = atomic.LoadInt32(&af.value) == 1
	return
}

// AtomicCounter is a counter to help with atomic operations.
type AtomicCounter struct {
	value int32
}

// Increment the value.
func (ac *AtomicCounter) Increment() {
	atomic.AddInt32(&ac.value, 1)
}

// Decrement the value.
func (ac *AtomicCounter) Decrement() {
	atomic.AddInt32(&ac.value, -1)
}

// Get returns the counter value.
func (ac *AtomicCounter) Get() (value int32) {
	value = atomic.LoadInt32(&ac.value)
	return
}
