package worker

import (
	"sync"
	"sync/atomic"
)

// Latch is a helper to coordinate killing goroutines.
// The lifecycle is generally as follows.
// 0 - stopped
// 1 - signal started
// 2 - running / started
// N-2 - signal stop
// N-1 - stopping
// goto 0
type Latch struct {
	starting int32
	running  int32
	stopping int32

	signalLock sync.Mutex
	started    chan struct{}
	shouldStop chan struct{}
	stopped    chan struct{}
}

// IsStopped returns if the latch is stopped.
func (l *Latch) IsStopped() bool {
	return atomic.LoadInt32(&l.starting) == 0 &&
		atomic.LoadInt32(&l.running) == 0 &&
		atomic.LoadInt32(&l.stopping) == 0
}

// IsStarting indicates the latch is waiting to be scheduled.
func (l *Latch) IsStarting() bool {
	return atomic.LoadInt32(&l.starting) == 1
}

// IsRunning indicates we can signal to stop.
func (l *Latch) IsRunning() bool {
	return atomic.LoadInt32(&l.running) == 1
}

// IsStopping returns if the latch is waiting to finish stopping.
func (l *Latch) IsStopping() bool {
	return atomic.LoadInt32(&l.stopping) == 1
}

// NotifyStarted returns the started signal.
// It is used to coordinate the transition from starting -> started.
func (l *Latch) NotifyStarted() <-chan struct{} {
	return l.started
}

// NotifyStop returns the should stop signal.
// It is used to trigger the transition from running -> stopping -> stopped.
func (l *Latch) NotifyStop() <-chan struct{} {
	return l.shouldStop
}

// NotifyStopped returns the stopped signal.
// It is used to coordinate the transition from stopping -> stopped.
func (l *Latch) NotifyStopped() <-chan struct{} {
	return l.stopped
}

// Starting signals the latch is starting.
// This is typically done before you kick off a goroutine.
func (l *Latch) Starting() {
	l.signalLock.Lock()
	defer l.signalLock.Unlock()

	if !l.IsStopped() {
		return
	}
	atomic.StoreInt32(&l.starting, 1)
	atomic.StoreInt32(&l.running, 0)
	atomic.StoreInt32(&l.stopping, 0)
	l.started = make(chan struct{})
}

// Started signals that the latch is started and has entered
// the `IsRunning` state.
func (l *Latch) Started() {
	l.signalLock.Lock()
	defer l.signalLock.Unlock()

	if !l.IsStarting() {
		return
	}

	atomic.StoreInt32(&l.starting, 0)
	atomic.StoreInt32(&l.running, 1)
	atomic.StoreInt32(&l.stopping, 0)
	l.shouldStop = make(chan struct{})
	close(l.started)
}

// Stop signals the latch to stop.
// It could also be thought of as `SignalStopping`.
func (l *Latch) Stop() {
	l.signalLock.Lock()
	defer l.signalLock.Unlock()

	if !l.IsRunning() {
		return
	}

	atomic.StoreInt32(&l.starting, 0)
	atomic.StoreInt32(&l.running, 0)
	atomic.StoreInt32(&l.stopping, 1)

	l.stopped = make(chan struct{})
	close(l.shouldStop)
}

// Stopped signals the latch has stopped.
func (l *Latch) Stopped() {
	l.signalLock.Lock()
	defer l.signalLock.Unlock()

	if !l.IsStopping() {
		return
	}

	atomic.StoreInt32(&l.starting, 0)
	atomic.StoreInt32(&l.running, 0)
	atomic.StoreInt32(&l.stopping, 0)
	close(l.stopped)
}
