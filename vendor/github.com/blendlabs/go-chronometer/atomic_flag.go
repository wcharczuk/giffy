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
	af.value = value
	af.lock.Unlock()
}

// Get the flag value.
func (af *AtomicFlag) Get() (value bool) {
	af.lock.RLock()
	value = af.value
	af.lock.RUnlock()
	return
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
	ac.value++
	ac.lock.Unlock()
}

// Decrement the value.
func (ac *AtomicCounter) Decrement() {
	ac.lock.Lock()
	ac.value--
	ac.lock.Unlock()
}

// Get returns the counter value.
func (ac *AtomicCounter) Get() (value int) {
	ac.lock.RLock()
	value = ac.value
	ac.lock.RUnlock()
	return
}
