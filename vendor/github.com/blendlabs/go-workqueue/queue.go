package workQueue

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"

	"github.com/blendlabs/go-exception"
)

const (
	// DefaultMaxRetries is the maximum times a process queue item will be retried before being dropped.
	DefaultMaxRetries = 10

	// DefaultMaxWorkItems is the default entry buffer length.
	DefaultMaxWorkItems = 1 << 10
)

var (
	_default     *Queue
	_defaultLock sync.Mutex
)

// Default returns a singleton queue.
func Default() *Queue {
	if _default == nil {
		_defaultLock.Lock()
		defer _defaultLock.Unlock()
		if _default == nil {
			_default = NewQueue()
		}
	}
	return _default
}

// Action is an action that can be dispatched by the process queue.
type Action func(args ...interface{}) error

// NewQueue returns a new work queue.
func NewQueue() *Queue {
	return &Queue{numWorkers: runtime.NumCPU(), maxRetries: DefaultMaxRetries, maxWorkItems: DefaultMaxWorkItems}
}

// NewQueueWithWorkers returns a new work queue with a given number of workers.
func NewQueueWithWorkers(numWorkers int) *Queue {
	return &Queue{numWorkers: numWorkers, maxRetries: DefaultMaxRetries, maxWorkItems: DefaultMaxWorkItems}
}

// NewQueueWithRetryCount returns a new work queue with a custom retry count.
func NewQueueWithRetryCount(retryCount int) *Queue {
	return &Queue{maxRetries: retryCount, maxWorkItems: DefaultMaxWorkItems}
}

// NewQueueWithMaxWorkItems returns a queue with a given maximum work queue length.
func NewQueueWithMaxWorkItems(workItems int) *Queue {
	return &Queue{numWorkers: runtime.NumCPU(), maxRetries: DefaultMaxRetries, maxWorkItems: workItems}
}

// Queue is the container for work items, it dispatches work to the workers.
type Queue struct {
	maxRetries   int
	maxWorkItems int

	actionQueue chan Entry

	synchronousDispatch bool
	numWorkers          int
	workers             []*Worker
	abortSignal         chan bool
}

// Start starts the dispatcher workers for the process quere.
func (q *Queue) Start() {
	q.workers = make([]*Worker, q.numWorkers)
	q.actionQueue = make(chan Entry, q.maxWorkItems)
	q.abortSignal = make(chan bool)

	for id := 0; id < q.numWorkers; id++ {
		q.newWorker(id)
	}

	q.dispatch()
}

// SetMaxWorkItems sets the max work items.
// Note: It MUST be called before .Start().
func (q *Queue) SetMaxWorkItems(workItems int) {
	q.maxWorkItems = workItems
	q.actionQueue = make(chan Entry, q.maxWorkItems)
}

// UseSynchronousDispatch sets the dispatcher to queue items synchronously (i.e. deterministically in order).
func (q *Queue) UseSynchronousDispatch() {
	q.synchronousDispatch = true
}

// UseAsyncDispatch sets the dispatcher to queue items asynchronously.
func (q *Queue) UseAsyncDispatch() {
	q.synchronousDispatch = false
}

// Len returns the number of items in the work queue.
func (q *Queue) Len() int {
	total := len(q.actionQueue)
	for _, w := range q.workers {
		total += len(w.WorkItems)
	}
	return total
}

// MaxRetries returns the maximum number of retries.
func (q *Queue) MaxRetries() int {
	return q.maxRetries
}

// Enqueue adds a work item to the process queue.
func (q *Queue) Enqueue(action Action, args ...interface{}) error {
	if q.actionQueue == nil {
		return exception.New("Work Queue has not been initialized; make sure to call `Start(...)` first.")
	}
	if len(q.actionQueue) > q.maxWorkItems {
		return exception.New("Work Queue is full, cannot enqueue.")
	}
	q.actionQueue <- Entry{Action: action, Args: args}
	return nil
}

// Drain drains the queue and stops the workers.
func (q *Queue) Drain() error {
	for x := 0; x < len(q.workers); x++ {
		q.workers[x].Stop()
	}
	q.abortSignal <- true
	q.workers = nil
	q.actionQueue = nil
	return nil
}

// String returns a string representation of the queue.
func (q *Queue) String() string {
	b := bytes.NewBuffer([]byte{})
	b.WriteString(fmt.Sprintf("WorkQueue [%d]", q.Len()))
	if q.Len() > 0 {
		q.VisitEach(func(e Entry) {
			b.WriteString(" ")
			b.WriteString(e.String())
		})
	}
	return b.String()
}

// VisitEach runs the consumer for each item in the queue.
// Useful for printing the actions in the queue.
func (q *Queue) VisitEach(visitor func(entry Entry)) {
	queueLength := len(q.actionQueue)
	var entry Entry
	for x := 0; x < queueLength; x++ {
		entry = <-q.actionQueue
		visitor(entry)
		q.actionQueue <- entry
	}
}

func (q *Queue) newWorker(id int) {
	q.workers[id] = NewWorker(id, q, q.maxWorkItems)
	q.workers[id].Start()
}

func (q *Queue) dispatch() {
	go func() {
		var workerIndex int
		for {
			select {
			case workItem := <-q.actionQueue:
				if q.synchronousDispatch {
					q.workers[workerIndex].WorkItems <- workItem
				} else {
					go func(worker int) { q.workers[worker].WorkItems <- workItem }(workerIndex)
				}
			case <-q.abortSignal:
				return
			}

			if q.numWorkers > 1 {
				workerIndex++
				if workerIndex >= q.numWorkers {
					workerIndex = 0
				}
			}
		}
	}()
}
