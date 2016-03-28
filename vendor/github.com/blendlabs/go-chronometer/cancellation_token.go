package chronometer

import "github.com/blendlabs/go-exception"

// NewCancellationToken returns a new CancellationToken instance.
func NewCancellationToken() *CancellationToken {
	return &CancellationToken{shouldCancel: false, didCancel: false, cancellationSignal: make(chan bool, 1)}
}

// CancellationException is the panic that gets raised when tasks are canceled.
type CancellationException error

// NewCancellationException returns a new cancellation exception.
func NewCancellationException() error {
	return CancellationException(exception.New("Cancellation grace period expired."))
}

// CancellationToken are the signalling mechanism chronometer uses to tell tasks that they should stop work.
type CancellationToken struct {
	shouldCancel       bool
	didCancel          bool
	cancellationSignal chan bool
}

func (ct *CancellationToken) signalCancellation() {
	ct.shouldCancel = true
	ct.didCancel = false
}

// ShouldCancel indicates if a token should cancel.
func (ct *CancellationToken) ShouldCancel() bool {
	return ct.shouldCancel
}

// Cancel should be called when ShouldCancel is true in order to signal OnCancellationReceiver's.
func (ct *CancellationToken) Cancel() error {
	ct.didCancel = true
	return exception.New("Task Cancellation")
}
