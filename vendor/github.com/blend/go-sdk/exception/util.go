package exception

// Is is a helper function that returns if an error is an exception.
func Is(err error) bool {
	if _, typedOk := err.(*Ex); typedOk {
		return true
	}
	return false
}

// As is a helper method that returns an error as an exception.
func As(err error) *Ex {
	if typed, typedOk := err.(*Ex); typedOk {
		return typed
	}
	return nil
}
