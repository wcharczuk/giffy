package workQueue

// NewWorker creates a new worker.
func NewWorker(id int, parent *Queue, maxWorkItems int) *Worker {
	return &Worker{
		ID:           id,
		MaxWorkItems: maxWorkItems,
		WorkItems:    make(chan Entry, maxWorkItems),
		Parent:       parent,
		Abort:        make(chan bool),
	}
}

// Worker is a consumer of the work queue.
type Worker struct {
	ID           int
	MaxWorkItems int
	WorkItems    chan Entry
	Parent       *Queue
	Abort        chan bool
}

// Start starts the worker.
func (w *Worker) Start() {
	go func() {
		var err error
		for {
			select {
			case workItem := <-w.WorkItems:
				err = workItem.Action(workItem.Args...)
				if err != nil {
					workItem.Tries++
					if workItem.Tries < w.Parent.maxRetries {
						w.Parent.actionQueue <- workItem
					}
				}
			case <-w.Abort:
				return
			}
		}
	}()
}

// Stop sends the stop signal to the worker.
func (w *Worker) Stop() {
	go func() {
		w.Abort <- true
	}()
}
