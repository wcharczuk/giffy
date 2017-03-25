package web

import "sync"

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

// Add a session to the cache.
func (sc *SessionCache) Add(session *Session) {
	sc.SessionLock.Lock()
	sc.Sessions[session.SessionID] = session
	sc.SessionLock.Unlock()
}

// Expire removes a session from the cache.
func (sc *SessionCache) Expire(sessionID string) {
	sc.SessionLock.Lock()
	delete(sc.Sessions, sessionID)
	sc.SessionLock.Unlock()
}

// IsActive returns if a sessionID is active.
func (sc *SessionCache) IsActive(sessionID string) bool {
	sc.SessionLock.Lock()
	_, hasSession := sc.Sessions[sessionID]
	sc.SessionLock.Unlock()
	return hasSession
}

// Get gets a session.
func (sc *SessionCache) Get(sessionID string) (*Session, bool) {
	sc.SessionLock.Lock()
	defer sc.SessionLock.Unlock()
	if session, hasSession := sc.Sessions[sessionID]; hasSession {
		return session, hasSession
	}
	return nil, false
}
