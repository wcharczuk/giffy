package chronometer

import "sync"

// AtomicFlag is a boolean value that is syncronized.
type AtomicFlag struct {
	lock  sync.RWMutex
	value bool
}

// Set the flag value.
func (af *AtomicFlag) Set(value bool) {
	af.lock.Lock()
	defer af.lock.Unlock()
	af.value = value
}

// Get the flag value.
func (af *AtomicFlag) Get() bool {
	af.lock.RLock()
	defer af.lock.RUnlock()
	return af.value
}

// AtomicCounter is a counter that you should use
// instead of atomic.AddInt32(...)
type AtomicCounter struct {
	lock  sync.RWMutex
	value int
}

// Increment the value.
func (ac *AtomicCounter) Increment() {
	ac.lock.Lock()
	defer ac.lock.Unlock()
	ac.value++
}

// Decrement the value.
func (ac *AtomicCounter) Decrement() {
	ac.lock.Lock()
	defer ac.lock.Unlock()
	ac.value--
}

// Get returns the counter value.
func (ac *AtomicCounter) Get() int {
	ac.lock.RLock()
	defer ac.lock.RUnlock()
	return ac.value
}
