package collections

// Queue is an interface for implementations of a FIFO buffer.
type Queue interface {
	Len() int
	Enqueue(value interface{})
	Dequeue() interface{}
	Peek() interface{}
	PeekBack() interface{}
	AsSlice() []interface{}
	Clear()
}
