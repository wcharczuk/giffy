package model

import (
	"fmt"
	"net/http"
	"time"

	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/uuid"
)

// NewError creates a new error.
func NewError(err error, req *http.Request) *Error {
	if _, isException := err.(*exception.Ex); isException {
		return &Error{
			UUID:       uuid.V4().String(),
			CreatedUTC: time.Now().UTC(),
			Message:    fmt.Sprintf("%v", err),
			StackTrace: fmt.Sprintf("%+v", err),

			Verb:  req.Method,
			Proto: req.Proto,
			Host:  req.Host,
			Path:  req.URL.Path,
			Query: req.URL.RawQuery,
		}
	}
	return &Error{
		UUID:       uuid.V4().String(),
		CreatedUTC: time.Now().UTC(),
		Message:    err.Error(),

		Verb:  req.Method,
		Proto: req.Proto,
		Host:  req.Host,
		Path:  req.URL.Path,
		Query: req.URL.RawQuery,
	}
}

// Error represents an exception that has bubbled up to the global exception handler.
type Error struct {
	UUID       string    `json:"uuid" db:"uuid,pk"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	Message    string    `json:"message" db:"message"`
	StackTrace string    `json:"stack_trace" db:"stack_trace"`

	Verb  string `json:"verb" db:"verb"`
	Proto string `json:"proto" db:"proto"`
	Host  string `json:"host" db:"host"`
	Path  string `json:"path" db:"path"`
	Query string `json:"query" db:"query"`
}

// TableName returns the mapped table name.
func (e Error) TableName() string {
	return "error"
}
