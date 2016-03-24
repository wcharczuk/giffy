package collections

import "strings"

// StringSet is a set of strings
type StringSet map[string]bool

// Add adds an element.
func (ss StringSet) Add(entry string) {
	if _, hasEntry := ss[entry]; !hasEntry {
		ss[entry] = true
	}
}

// Contains returns if an element is in the set.
func (ss StringSet) Contains(entry string) bool {
	_, hasEntry := ss[entry]
	return hasEntry
}

// Remove deletes an element, returns if the element was in the set.
func (ss StringSet) Remove(entry string) bool {
	if _, hasEntry := ss[entry]; hasEntry {
		delete(ss, entry)
		return true
	}
	return false
}

// Length returns the length of the set.
func (ss StringSet) Length() int {
	return len(ss)
}

// Copy returns a new copy of the set.
func (ss StringSet) Copy() StringSet {
	newSet := StringSet{}
	for key := range ss {
		newSet.Add(key)
	}
	return newSet
}

// ToArray returns the set as an array.
func (ss StringSet) ToArray() []string {
	output := []string{}
	for key := range ss {
		output = append(output, key)
	}
	return output
}

// String returns the set as a csv string.
func (ss StringSet) String() string {
	return strings.Join(ss.ToArray(), ", ")
}
