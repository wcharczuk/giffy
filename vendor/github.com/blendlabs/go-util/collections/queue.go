package collections

type queueNode struct {
	Next     *queueNode
	Previous *queueNode
	Value    interface{}
}

// NewQueue returns a new Queue instance.
func NewQueue() *Queue {
	return &Queue{}
}

// Queue is an implementation of a FIFO buffer.
// Remarks; it is not threadsafe. It is constant time in all ops.
type Queue struct {
	head   *queueNode
	tail   *queueNode
	length int
}

// Length returns the length of the queue in constant time.
func (q *Queue) Length() int {
	return q.length
}

// Push adds a new value to the queue.
func (q *Queue) Push(value interface{}) {
	node := queueNode{Value: value}

	if q.head == nil { //the queue is empty, that is to say head is nil
		q.head = &node
		q.tail = &node
	} else { //the queue is not empty, we have a (valid) tail pointer
		q.tail.Previous = &node
		node.Next = q.tail
		q.tail = &node
	}

	q.length = q.length + 1
}

// Dequeue removes an item from the front of the queue and returns it.
func (q *Queue) Dequeue() interface{} {
	if q.head == nil {
		return nil
	}

	headValue := q.head.Value

	if q.length == 1 && q.head == q.tail {
		q.head = nil
		q.tail = nil
	} else {
		q.head = q.head.Previous
		if q.head != nil {
			q.head.Next = nil
		}
	}

	q.length = q.length - 1
	return headValue
}

// Peek returns the first element of the queue but does not remove it.
func (q *Queue) Peek() interface{} {
	if q.head == nil {
		return nil
	}
	return q.head.Value
}

// PeekBack returns the last element of the queue.
func (q *Queue) PeekBack() interface{} {
	if q.tail == nil {
		return nil
	}
	return q.tail.Value
}

// ToArray returns the full contents of the queue as a slice.
func (q *Queue) ToArray() []interface{} {
	if q.head == nil {
		return []interface{}{}
	}

	values := []interface{}{}
	nodePtr := q.head
	for nodePtr != nil {
		values = append(values, nodePtr.Value)
		nodePtr = nodePtr.Previous
	}
	return values
}
