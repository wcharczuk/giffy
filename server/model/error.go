package model

import (
	"net/http"
	"time"

	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/uuid"
)

// NewError creates a new error.
func NewError(err error, req *http.Request) *Error {
	var merr Error
	if typed, ok := err.(*exception.Ex); ok {
		merr = Error{
			UUID:       uuid.V4().String(),
			CreatedUTC: time.Now().UTC(),
			Message:    typed.Message(),
			StackTrace: typed.Stack().String(),
		}
	} else {
		merr = Error{
			UUID:       uuid.V4().String(),
			CreatedUTC: time.Now().UTC(),
			Message:    err.Error(),
		}
	}
	if req != nil {
		merr.Verb = req.Method
		merr.Proto = req.Proto
		merr.Host = req.Host
		merr.Path = req.URL.Path
		merr.Query = req.URL.RawQuery
	}
	return &merr
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
