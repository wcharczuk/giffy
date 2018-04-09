package workqueue

import "sync/atomic"

// NewWorker creates a new worker.
func NewWorker(id int, parent *Queue, maxItems int) *Worker {
	return &Worker{
		ID:     id,
		Parent: parent,

		Work:    make(chan *Entry, maxItems),
		Abort:   make(chan bool),
		Aborted: make(chan bool),
	}
}

// Worker is a consumer of the work queue.
type Worker struct {
	ID      int
	Work    chan *Entry
	Parent  *Queue
	Abort   chan bool
	Aborted chan bool
}

// Start starts the worker.
func (w *Worker) Start() {
	go w.processWork()
}

func (w *Worker) processWork() {
	var err error
	var workItem *Entry
	for {
		select {
		case workItem = <-w.Work:
			if workItem == nil {
				continue
			}
			err = workItem.Execute()
			if err != nil {
				atomic.AddInt32(&workItem.Tries, 1)
				if workItem.Tries < int32(w.Parent.maxRetries) {
					w.Parent.work <- workItem
					continue
				}
			}
			w.Parent.entryPool.Put(workItem)
		case <-w.Abort:
			w.Aborted <- true
			return
		}
	}
}

// Close sends the stop signal to the worker.
func (w *Worker) Close() error {
	w.Abort <- true
	<-w.Aborted
	close(w.Abort)
	close(w.Aborted)
	close(w.Work)

	w.Abort = nil
	w.Aborted = nil
	w.Work = nil
	w.Parent = nil
	return nil
}
