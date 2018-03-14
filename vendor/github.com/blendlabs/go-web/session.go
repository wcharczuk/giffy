package web

import (
	"sync"
	"time"
)

// NewSession returns a new session object.
func NewSession(userID string, sessionID string) *Session {
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
		State:      map[string]interface{}{},
		Mutex:      &sync.RWMutex{},
	}
}

// Session is an active session
type Session struct {
	UserID     string                 `json:"userID" yaml:"userID"`
	SessionID  string                 `json:"sessionID" yaml:"sessionID"`
	CreatedUTC time.Time              `json:"createdUTC" yaml:"createdUTC"`
	ExpiresUTC *time.Time             `json:"expiresUTC" yaml:"expiresUTC"`
	State      map[string]interface{} `json:"state,omitempty" yaml:"state,omitempty"`
	Mutex      *sync.RWMutex          `json:"-" yaml:"-"`
}

// IsExpired returns if the session is expired.
func (s *Session) IsExpired() bool {
	if s == nil {
		return false
	}
	if s.ExpiresUTC == nil {
		return false
	}
	return s.ExpiresUTC.Before(time.Now().UTC())
}

func (s *Session) ensureMutex() {
	if s.Mutex == nil {
		s.Mutex = &sync.RWMutex{}
	}
}

// IsZero returns if the object is set or not.
// It will return true if either the userID or the sessionID are unset.
func (s *Session) IsZero() bool {
	if s == nil {
		return true
	}
	return len(s.UserID) == 0 || len(s.SessionID) == 0
}

// Lock locks the session.
func (s *Session) Lock() {
	s.ensureMutex()
	s.Mutex.Lock()
}

// Unlock unlocks the session.
func (s *Session) Unlock() {
	s.ensureMutex()
	s.Mutex.Unlock()
}

// RLock read locks the session.
func (s *Session) RLock() {
	s.ensureMutex()
	s.Mutex.RLock()
}

// RUnlock read unlocks the session.
func (s *Session) RUnlock() {
	s.ensureMutex()
	s.Mutex.RUnlock()
}
