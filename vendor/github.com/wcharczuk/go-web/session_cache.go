package web

// NewSessionCache returns a new session cache.
func NewSessionCache() *SessionCache {
	return &SessionCache{
		Sessions: map[string]*Session{},
	}
}

// SessionCache is a memory ledger of active sessions.
type SessionCache struct {
	Sessions map[string]*Session
}

// Add a session to the cache.
func (sc *SessionCache) Add(session *Session) {
	sc.Sessions[session.SessionID] = session
}

// Expire removes a session from the cache.
func (sc *SessionCache) Expire(sessionID string) {
	delete(sc.Sessions, sessionID)
}

// IsActive returns if a sessionID is active.
func (sc *SessionCache) IsActive(sessionID string) bool {
	_, hasSession := sc.Sessions[sessionID]
	return hasSession
}

// Get gets a session.
func (sc *SessionCache) Get(sessionID string) *Session {
	return sc.Sessions[sessionID]
}
