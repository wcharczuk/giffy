package web

import (
	"sync"
	"time"
)

// NewSession returns a new session object.
func NewSession(userID int64, sessionID string) *Session {
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
		State:      map[string]interface{}{},
		lock:       &sync.RWMutex{},
	}
}

// Session is an active session
type Session struct {
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	CreatedUTC time.Time              `json:"created_utc"`
	State      map[string]interface{} `json:"-"`
	lock       *sync.RWMutex
}

// Lock locks the session.
func (s *Session) Lock() {
	s.lock.Lock()
}

// Unlock unlocks the session.
func (s *Session) Unlock() {
	s.lock.Unlock()
}

// RLock read locks the session.
func (s *Session) RLock() {
	s.lock.RLock()
}

// RUnlock read unlocks the session.
func (s *Session) RUnlock() {
	s.lock.RUnlock()
}
