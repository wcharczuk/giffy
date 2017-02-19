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
	UserID     int64
	SessionID  string
	CreatedUTC time.Time
	State      map[string]interface{}
	lock       *sync.RWMutex
}

func (s *Session) ensureLock() {
	if s.lock == nil {
		s.lock = &sync.RWMutex{}
	}
}

// IsZero returns if the object is set or not.
func (s *Session) IsZero() bool {
	return s.UserID == 0 || len(s.SessionID) == 0
}

// Lock locks the session.
func (s *Session) Lock() {
	s.ensureLock()
	s.lock.Lock()
}

// Unlock unlocks the session.
func (s *Session) Unlock() {
	s.ensureLock()
	s.lock.Unlock()
}

// RLock read locks the session.
func (s *Session) RLock() {
	s.ensureLock()
	s.lock.RLock()
}

// RUnlock read unlocks the session.
func (s *Session) RUnlock() {
	s.ensureLock()
	s.lock.RUnlock()
}
