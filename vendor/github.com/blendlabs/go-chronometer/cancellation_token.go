package chronometer

import "github.com/blendlabs/go-exception"

// NewCancellationToken returns a new CancellationToken instance.
func NewCancellationToken() *CancellationToken {
	return &CancellationToken{ShouldCancel: false, didCancel: false}
}

// CancellationToken are the signalling mechanism chronometer uses to tell tasks that they should stop work.
type CancellationToken struct {
	ShouldCancel bool
	didCancel    bool
}

func (ct *CancellationToken) signalCancellation() {
	ct.ShouldCancel = true
	ct.didCancel = false
}

// Cancel should be called when ShouldCancel is true in order to signal OnCancellationReceiver's.
func (ct *CancellationToken) Cancel() error {
	ct.didCancel = true
	return exception.New("Task Cancellation")
}
