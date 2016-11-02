package core

// NewSetOfInt64 returns a new SetOfInt64
func NewSetOfInt64(values []int64) SetOfInt64 {
	set := SetOfInt64{}
	for _, v := range values {
		set.Add(v)
	}
	return set
}

// SetOfInt64 is a type alias for map[int]int
type SetOfInt64 map[int64]bool

// Add adds an element to the set, replaceing a previous value.
func (is SetOfInt64) Add(i int64) {
	is[i] = true
}

// Remove removes an element from the set.
func (is SetOfInt64) Remove(i int64) {
	delete(is, i)
}

// Contains returns if the element is in the set.
func (is SetOfInt64) Contains(i int64) bool {
	_, ok := is[i]
	return ok
}
