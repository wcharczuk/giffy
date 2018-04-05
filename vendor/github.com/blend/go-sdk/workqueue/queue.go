package workqueue

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
)

const (
	// DefaultMaxRetries is the maximum times a process queue item will be retried before being dropped.
	DefaultMaxRetries = 10

	// DefaultMaxWorkItems is the default entry buffer length.
	// Currently the default is 2^18 or 256k.
	// WorkItems maps to the initialized capacity of a buffered channel.
	// As a result it does not reflect actual memory consumed.
	DefaultMaxWorkItems = 1 << 18
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
			_default = New()
		}
	}
	return _default
}

// Action is an action that can be dispatched by the process queue.
type Action func(args ...interface{}) error

// New returns a new work queue.
func New() *Queue {
	return &Queue{
		recover:      true,
		numWorkers:   runtime.NumCPU(),
		maxRetries:   DefaultMaxRetries,
		maxWorkItems: DefaultMaxWorkItems,
	}
}

// NewWithWorkers returns a new work queue with a given number of workers.
func NewWithWorkers(numWorkers int) *Queue {
	return &Queue{
		recover:      true,
		numWorkers:   numWorkers,
		maxRetries:   DefaultMaxRetries,
		maxWorkItems: DefaultMaxWorkItems,
	}
}

// NewWithOptions returns a new queue with customizable options.
func NewWithOptions(numWorkers, retryCount, maxWorkItems int) *Queue {
	return &Queue{
		recover:      true,
		numWorkers:   numWorkers,
		maxRetries:   retryCount,
		maxWorkItems: maxWorkItems,
	}
}

// Queue is the container for work items, it dispatches work to the workers.
type Queue struct {
	numWorkers   int
	maxRetries   int
	maxWorkItems int

	running bool
	recover bool

	work chan *Entry

	entryPool sync.Pool
	workers   []*Worker
	abort     chan bool
	aborted   chan bool
}

// Start starts the dispatcher workers for the process quere.
func (q *Queue) Start() {
	if q.running {
		return
	}

	q.workers = make([]*Worker, q.numWorkers)
	q.work = make(chan *Entry, q.maxWorkItems)
	q.abort = make(chan bool)
	q.aborted = make(chan bool)
	q.entryPool = sync.Pool{
		New: func() interface{} {
			return &Entry{}
		},
	}
	q.running = true

	for id := 0; id < q.numWorkers; id++ {
		q.createAndStartWorker(id)
	}

	go q.dispatch()
}

// Recover returns if the queue is handling / recovering from panics.
func (q *Queue) Recover() bool {
	return q.recover
}

// SetRecover sets if the queue workers should handle panics.
func (q *Queue) SetRecover(shouldRecover bool) {
	q.recover = shouldRecover
}

// WithRecover sets if the queue should recover panics.
func (q *Queue) WithRecover(shouldRecover bool) *Queue {
	q.recover = shouldRecover
	return q
}

// Len returns the number of items in the work queue.
func (q *Queue) Len() int {
	return len(q.work)
}

// NumWorkers returns the number of worker routines.
func (q *Queue) NumWorkers() int {
	return q.numWorkers
}

// SetNumWorkers lets you set the num workers.
func (q *Queue) SetNumWorkers(workers int) {
	q.numWorkers = workers
	if q.running {
		q.Close()
		q.Start()
	}
}

// WithNumWorkers calls `SetNumWorkers` and returns a reference to the queue.
func (q *Queue) WithNumWorkers(workers int) *Queue {
	q.SetNumWorkers(workers)
	return q
}

// MaxWorkItems returns the maximum length of the work item queue.
func (q *Queue) MaxWorkItems() int {
	return q.maxWorkItems
}

// SetMaxWorkItems sets the max work items.
func (q *Queue) SetMaxWorkItems(workItems int) {
	q.maxWorkItems = workItems
	if q.running {
		q.Close()
		q.Start()
	}
}

// WithMaxWorkItems calls `SetMaxWorkItems` and returns a reference to the queue.
func (q *Queue) WithMaxWorkItems(workItems int) *Queue {
	q.SetMaxWorkItems(workItems)
	return q
}

// MaxRetries returns the maximum number of retries.
func (q *Queue) MaxRetries() int {
	return q.maxRetries
}

// SetMaxRetries sets the maximum nummer of retries for a work item on error.
func (q *Queue) SetMaxRetries(maxRetries int) {
	q.maxRetries = maxRetries
}

// WithMaxRetries calls `SetMaxRetries` and returns a reference to the queue.
func (q *Queue) WithMaxRetries(maxRetries int) *Queue {
	q.SetMaxRetries(maxRetries)
	return q
}

// Running returns if the queue has started or not.
func (q *Queue) Running() bool {
	return q.running
}

// Enqueue adds a work item to the process queue.
func (q *Queue) Enqueue(action Action, args ...interface{}) {
	if !q.running {
		return
	}
	entry := q.entryPool.Get().(*Entry)
	entry.Recover = q.recover
	entry.Action = action
	entry.Args = args
	entry.Tries = 0
	q.work <- entry
}

// Close drains the queue and stops the workers.
func (q *Queue) Close() error {
	if !q.running {
		return nil
	}

	q.abort <- true
	<-q.aborted

	close(q.abort)
	close(q.aborted)
	close(q.work)

	var err error
	for x := 0; x < len(q.workers); x++ {
		err = q.workers[x].Close()
		if err != nil {
			return err
		}
	}

	q.workers = nil
	q.work = nil
	q.abort = nil
	q.aborted = nil
	q.running = false
	return nil
}

// String returns a string representation of the queue.
func (q *Queue) String() string {
	b := bytes.NewBuffer([]byte{})
	b.WriteString(fmt.Sprintf("WorkQueue [%d]", q.Len()))
	if q.Len() > 0 {
		q.Each(func(e *Entry) {
			b.WriteString(" ")
			b.WriteString(e.String())
		})
	}
	return b.String()
}

// Each runs the consumer for each item in the queue.
func (q *Queue) Each(visitor func(entry *Entry)) {
	queueLength := len(q.work)
	var entry *Entry
	for x := 0; x < queueLength; x++ {
		entry = <-q.work
		visitor(entry)
		q.work <- entry
	}
}

func (q *Queue) createAndStartWorker(id int) {
	q.workers[id] = NewWorker(id, q, q.maxWorkItems/q.numWorkers)
	q.workers[id].Start()
}

func (q *Queue) dispatch() {
	var workItem *Entry
	var workerIndex int
	numWorkers := len(q.workers)
	for {
		select {
		case workItem = <-q.work:
			if workItem == nil {
				continue
			}
			q.workers[workerIndex].Work <- workItem
			if numWorkers > 1 {
				workerIndex++
				if workerIndex >= numWorkers {
					workerIndex = 0
				}
			}
		case <-q.abort:
			q.aborted <- true
			return
		}
	}
}
