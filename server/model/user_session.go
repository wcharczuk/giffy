package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
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

// DeleteUserSession removes a session from the db.
func DeleteUserSession(userID int64, sessionID string, tx *sql.Tx) error {
	return spiffy.DefaultDb().ExecInTransaction("DELETE FROM user_session where user_id = $1 and session_id = $2", tx, userID, sessionID)
}
