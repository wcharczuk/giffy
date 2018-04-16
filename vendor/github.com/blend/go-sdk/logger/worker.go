package logger

import (
	"sync"
)

const (
	// DefaultWorkerQueueDepth is the default depth per listener to queue work.
	DefaultWorkerQueueDepth = 1 << 20
)

// NewWorker returns a new worker.
func NewWorker(parent *Logger, listener Listener) *Worker {
	return &Worker{
		Parent:        parent,
		Listener:      listener,
		Work:          make(chan Event, DefaultWorkerQueueDepth),
		Abort:         make(chan bool),
		Aborted:       make(chan bool),
		RecoverPanics: true,
	}
}

// Worker is an agent that processes a listener.
type Worker struct {
	Parent        *Logger
	Listener      Listener
	Abort         chan bool
	Aborted       chan bool
	Drained       chan bool
	Work          chan Event
	SyncRoot      sync.Mutex
	RecoverPanics bool
}

// WithRecoverPanics sets the recover panics field.
func (w *Worker) WithRecoverPanics(value bool) *Worker {
	w.RecoverPanics = value
	return w
}

// Start starts the worker.
func (w *Worker) Start() {
	go w.ProcessLoop()
}

// ProcessLoop is the for/select loop.
func (w *Worker) ProcessLoop() {
	var e Event
	for {
		select {
		case e = <-w.Work:
			w.Process(e)
		case <-w.Abort:
			w.Aborted <- true
			return
		}
	}
}

// Process calls the listener for an event.
func (w *Worker) Process(e Event) {
	if w.RecoverPanics {
		defer func() {
			if r := recover(); r != nil {
				if w.Parent != nil {
					w.Parent.SyncFatalf("%v", r)
				}
			}
		}()
	}

	w.Listener(e)
}

// Stop stops the worker.
func (w *Worker) Stop() {
	w.Abort <- true
	<-w.Aborted
}

// Drain stops the worker and synchronously processes any remaining work.
// It then restarts the worker.
func (w *Worker) Drain() {
	w.SyncRoot.Lock()
	defer w.SyncRoot.Unlock()

	w.Stop()
	for len(w.Work) > 0 {
		w.Process(<-w.Work)
	}
	w.Start()
}

// Close closes the worker.
func (w *Worker) Close() error {
	w.SyncRoot.Lock()
	defer w.SyncRoot.Unlock()

	w.Stop()

	close(w.Work)
	close(w.Abort)
	close(w.Aborted)

	return nil
}
