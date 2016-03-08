package model

import (
	"time"

	"github.com/blendlabs/go-util"
)

// UserSession is a session for a user
type UserSession struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	SessionID    string    `json:"session_id" db:"session_id,pk"`
}

// TableName returns the table name.
func (us UserSession) TableName() string {
	return "user_session"
}

// IsZero returns if a session is zero or not.
func (us UserSession) IsZero() bool {
	return us.UserID == 0 || len(us.SessionID) == 0
}

// NewUserSession returns a new user session.
func NewUserSession(userID int64) *UserSession {
	return &UserSession{
		UserID:       userID,
		TimestampUTC: time.Now().UTC(),
		SessionID:    util.RandomString(32),
	}
}
