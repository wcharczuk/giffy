package web

import (
	"sync"
)

// NewSessionCache returns a new session cache.
func NewSessionCache() *SessionCache {
	return &SessionCache{
		SessionLock: &sync.Mutex{},
		Sessions:    map[string]*Session{},
	}
}

// SessionCache is a memory ledger of active sessions.
type SessionCache struct {
	SessionLock *sync.Mutex
	Sessions    map[string]*Session
}

// Upsert adds or updates a session to the cache.
func (sc *SessionCache) Upsert(session *Session) {
	sc.SessionLock.Lock()
	defer sc.SessionLock.Unlock()
	sc.Sessions[session.SessionID] = session
}

// Remove removes a session from the cache.
func (sc *SessionCache) Remove(sessionID string) {
	sc.SessionLock.Lock()
	defer sc.SessionLock.Unlock()
	delete(sc.Sessions, sessionID)
}

// Get gets a session.
func (sc *SessionCache) Get(sessionID string) *Session {
	sc.SessionLock.Lock()
	defer sc.SessionLock.Unlock()

	if session, hasSession := sc.Sessions[sessionID]; hasSession {
		return session
	}
	return nil
}

// IsActive returns if a sessionID is active.
func (sc *SessionCache) IsActive(sessionID string) bool {
	sc.SessionLock.Lock()
	defer sc.SessionLock.Unlock()

	session, hasSession := sc.Sessions[sessionID]
	if hasSession {
		return !session.IsExpired()
	}
	return hasSession
}
