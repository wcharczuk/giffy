package util

var (
	// BitFlag is a namespace for bitflag functions.
	BitFlag = bitFlag{}
)

type bitFlag struct{}

// All returns if all the reference bits are set for a given value
func (bf bitFlag) All(reference, value int) bool {
	return reference&value == value
}

// Any returns if any the reference bits are set for a given value
func (bf bitFlag) Any(reference, value int) bool {
	return reference&value > 0
}

// Combine combines all the values into one flag.
func (bf bitFlag) Combine(values ...int) int {
	outputFlag := 0
	for _, value := range values {
		outputFlag = outputFlag | value
	}
	return outputFlag
}
