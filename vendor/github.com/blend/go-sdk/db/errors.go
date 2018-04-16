package db

import "github.com/blend/go-sdk/exception"

const (
	// ErrUnsafeSSLMode is an error indicating unsafe ssl mode in production.
	ErrUnsafeSSLMode Error = "db: unsafe ssl mode in prodlike environment"
	// ErrUsernameUnset is an error indicating there is no username set in a prodlike environment.
	ErrUsernameUnset Error = "db: username is unset in prodlike environment"
	// ErrPasswordUnset is an error indicating there is no password set in a prodlike environment.
	ErrPasswordUnset Error = "db: password is unset in prodlike environment"
)

// IsUnsafeSSLMode returns if an error is an `ErrUnsafeSSLMode`.
func IsUnsafeSSLMode(err error) bool {
	return exceptionInner(err) == ErrUnsafeSSLMode
}

// IsUsernameUnset returns if an error is an `ErrUsernameUnset`.
func IsUsernameUnset(err error) bool {
	return exceptionInner(err) == ErrUsernameUnset
}

// IsPasswordUnset returns if an error is an `ErrPasswordUnset`.
func IsPasswordUnset(err error) bool {
	return exceptionInner(err) == ErrPasswordUnset
}

func exceptionInner(err error) error {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err
}
