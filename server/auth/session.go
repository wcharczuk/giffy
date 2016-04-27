package auth

import (
	"sync"
	"time"

	"github.com/wcharczuk/giffy/server/model"
)

// NewSession returns a new session object.
func NewSession(userID int64, sessionID string) *Session {
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
		State:      map[string]interface{}{},
		lock:       sync.Mutex{},
	}
}

// Session is an active session
type Session struct {
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	CreatedUTC time.Time              `json:"created_utc"`
	User       *model.User            `json:"user"`
	State      map[string]interface{} `json:"-"`

	lock sync.Mutex
}

// Lock locks the session.
func (s *Session) Lock() {
	s.lock.Lock()
}

// Unlock unlocks the session.
func (s *Session) Unlock() {
	s.lock.Unlock()
}
