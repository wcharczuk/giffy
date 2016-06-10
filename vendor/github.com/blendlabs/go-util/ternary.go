package util

// Ternary returns one of the given values upon a bool.
func Ternary(condition bool, ifTrue, ifFalse interface{}) interface{} {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TernaryOfString returns one of the given values upon a bool.
func TernaryOfString(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TernaryOfInt returns one of the given values upon a bool.
func TernaryOfInt(condition bool, ifTrue, ifFalse int) int {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// TernaryOfFloat returns one of the given values upon a bool.
func TernaryOfFloat(condition bool, ifTrue, ifFalse float64) float64 {
	if condition {
		return ifTrue
	}
	return ifFalse
}
