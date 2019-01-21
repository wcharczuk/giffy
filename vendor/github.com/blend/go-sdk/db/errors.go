package db

import (
	"github.com/blend/go-sdk/exception"
	"github.com/lib/pq"
)

const (
	// ErrDestinationNotStruct is an exception class.
	ErrDestinationNotStruct exception.Class = "db: destination object is not a struct"
	// ErrConfigUnset is an exception class.
	ErrConfigUnset exception.Class = "db: config is unset"
	// ErrUnsafeSSLMode is an error indicating unsafe ssl mode in production.
	ErrUnsafeSSLMode exception.Class = "db: unsafe ssl mode in prodlike environment"
	// ErrUsernameUnset is an error indicating there is no username set in a prodlike environment.
	ErrUsernameUnset exception.Class = "db: username is unset in prodlike environment"
	// ErrPasswordUnset is an error indicating there is no password set in a prodlike environment.
	ErrPasswordUnset exception.Class = "db: password is unset in prodlike environment"
	// ErrConnectionAlreadyOpen is an error indicating the db connection was already opened.
	ErrConnectionAlreadyOpen exception.Class = "db: the connection is already opened"
	// ErrConnectionClosed is an error indicating the db connection hasn't been opened.
	ErrConnectionClosed exception.Class = "db: the connection is closed"
	// ErrPlanCacheUnset is an error indicating the statement cache is unset.
	ErrPlanCacheUnset exception.Class = "db: the plan cache is unset"
	// ErrPlanCacheKeyUnset is an error indicating the plan cache key is unset.
	ErrPlanCacheKeyUnset exception.Class = "db: the plan cache key is unset"
	// ErrCollectionNotSlice is an error returned by OutMany if the destination is not a slice.
	ErrCollectionNotSlice exception.Class = "db: outmany destination collection is not a slice"
	// ErrInvalidIDs is an error returned by Get if the ids aren't provided.
	ErrInvalidIDs exception.Class = "db: invalid `ids` parameter"
	// ErrNoPrimaryKey is an error returned by a number of operations that depend on a primary key.
	ErrNoPrimaryKey exception.Class = "db: no primary key on object"
	// ErrRowsNotColumnsProvider is returned by `PopulateByName` if you do not pass in `sql.Rows` as the scanner.
	ErrRowsNotColumnsProvider exception.Class = "db: rows is not a columns provider"
)

// IsConfigUnset returns if the error is an `ErrConfigUnset`.
func IsConfigUnset(err error) bool {
	return exception.Is(err, ErrConfigUnset)
}

// IsUnsafeSSLMode returns if an error is an `ErrUnsafeSSLMode`.
func IsUnsafeSSLMode(err error) bool {
	return exception.Is(err, ErrUnsafeSSLMode)
}

// IsUsernameUnset returns if an error is an `ErrUsernameUnset`.
func IsUsernameUnset(err error) bool {
	return exception.Is(err, ErrUsernameUnset)
}

// IsPasswordUnset returns if an error is an `ErrPasswordUnset`.
func IsPasswordUnset(err error) bool {
	return exception.Is(err, ErrPasswordUnset)
}

// IsConnectionClosed returns if the error is an `ErrConnectionClosed`.
func IsConnectionClosed(err error) bool {
	return exception.Is(err, ErrConnectionClosed)
}

// IsPlanCacheUnset returns if the error is an `ErrConnectionClosed`.
func IsPlanCacheUnset(err error) bool {
	return exception.Is(err, ErrPlanCacheUnset)
}

// IsPlanCacheKeyUnset returns if the error is an `ErrPlanCacheKeyUnset`.
func IsPlanCacheKeyUnset(err error) bool {
	return exception.Is(err, ErrPlanCacheKeyUnset)
}

// Error returns a new exception by parsing (potentially)
// a driver error into relevant pieces.
func Error(err error) error {
	if err == nil {
		return nil
	}
	if ex := exception.As(err); ex != nil {
		return ex
	}
	if typed, ok := err.(*pq.Error); ok {
		return exception.New(typed.Code.Name()).WithMessage(typed.Message).WithInner(err)
	}
	return exception.New(err)
}
