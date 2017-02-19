package chronometer

import (
	"sync"

	"github.com/blendlabs/go-exception"
)

// NewCancellationToken returns a new CancellationToken instance.
func NewCancellationToken() *CancellationToken {
	return &CancellationToken{
		shouldCancel:     false,
		shouldCancelLock: sync.RWMutex{},
	}
}

// CancellationPanic is the panic that gets raised when tasks are canceled.
type CancellationPanic error

// NewCancellationPanic returns a new cancellation exception.
func NewCancellationPanic() error {
	return CancellationPanic(exception.New("Cancellation grace period expired."))
}

// HandleCancellationPanic is a method to use in your
// execution handlers that handles and forwards the CancellationPanic
func HandleCancellationPanic(handler func()) {
	if r := recover(); r != nil {
		if _, isCancellation := r.(CancellationPanic); isCancellation {
			handler()
			panic(r)
		}
	}
}

// CancellationToken are the signalling mechanism chronometer uses to tell tasks that they should stop work.
type CancellationToken struct {
	shouldCancel     bool
	shouldCancelLock sync.RWMutex
}

// Cancel signals cancellation.
func (ct *CancellationToken) Cancel() {
	ct.shouldCancelLock.Lock()
	ct.shouldCancel = true
	ct.shouldCancelLock.Unlock()
}

func (ct *CancellationToken) didCancel() (value bool) {
	ct.shouldCancelLock.RLock()
	value = ct.shouldCancel
	ct.shouldCancelLock.RUnlock()
	return
}

// CheckCancellation indicates if a token has been signaled to cancel.
func (ct *CancellationToken) CheckCancellation() {
	ct.shouldCancelLock.RLock()
	defer ct.shouldCancelLock.RUnlock()
	if ct.shouldCancel {
		panic(NewCancellationPanic())
	}
}
