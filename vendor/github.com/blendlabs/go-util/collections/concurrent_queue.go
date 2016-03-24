package collections

import "sync"

// NewConcurrentQueue returns a new ConcurrentQueue instance.
func NewConcurrentQueue(maxSize int) *ConcurrentQueue {
	return &ConcurrentQueue{MaxSize: maxSize, storage: make(chan interface{}, maxSize), latch: sync.Mutex{}}
}

// ConcurrentQueue is a threadsafe queue.
type ConcurrentQueue struct {
	MaxSize int
	storage chan interface{}
	latch   sync.Mutex
}

// Length returns the number of items in the queue.
func (cq *ConcurrentQueue) Length() int {
	return len(cq.storage)
}

// Push adds an item to the queue.
func (cq *ConcurrentQueue) Push(item interface{}) {
	cq.storage <- item
}

// Dequeue returns the next element in the queue.
func (cq *ConcurrentQueue) Dequeue() interface{} {
	if len(cq.storage) != 0 {
		return <-cq.storage
	}
	return nil
}

// ToArray iterates over the queue and returns an array of its contents.
func (cq *ConcurrentQueue) ToArray() []interface{} {
	cq.latch.Lock()
	defer cq.latch.Unlock()

	values := []interface{}{}
	for len(cq.storage) != 0 {
		v := <-cq.storage
		values = append(values, v)
	}
	for _, v := range values {
		cq.storage <- v
	}
	return values
}
