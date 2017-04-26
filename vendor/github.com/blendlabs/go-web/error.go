package web

// Error is a simple wrapper for strings to help with constant errors.
type Error string

func (e Error) Error() string { return string(e) }
