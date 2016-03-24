package collections

// NewIntSet returns a new IntSet
func NewIntSet(values []int) IntSet {
	set := IntSet{}
	for _, v := range values {
		set.Add(v)
	}
	return set
}

// IntSet is a type alias for map[int]int
type IntSet map[int]bool

// Add adds an element to the set, replaceing a previous value.
func (is IntSet) Add(i int) {
	is[i] = true
}

// Remove removes an element from the set.
func (is IntSet) Remove(i int) {
	delete(is, i)
}

// Contains returns if the element is in the set.
func (is IntSet) Contains(i int) bool {
	_, ok := is[i]
	return ok
}
