package auth

import (
	"database/sql"
	"sync"

	"github.com/wcharczuk/giffy/server/model"
)

var (
	sessionCacheLatch = sync.Mutex{}
	sessionCache      *SessionCache
)

// SessionState returns the shared SessionCache singleton.
func SessionState() *SessionCache {
	if sessionCache == nil {
		sessionCacheLatch.Lock()
		defer sessionCacheLatch.Unlock()
		if sessionCache == nil {
			sessionCache = newSessionCache()
		}
	}

	return sessionCache
}

// NewSessionCache returns a new session cache.
func newSessionCache() *SessionCache {
	return &SessionCache{
		Sessions: map[string]*Session{},
	}
}

// SessionCache is a memory ledger of active sessions.
type SessionCache struct {
	Sessions map[string]*Session
}

// Add a session to the cache.
func (sc *SessionCache) Add(userID int64, sessionID string, tx *sql.Tx) (*Session, error) {
	session := NewSession(userID, sessionID)

	user, err := model.GetUserByID(session.UserID, tx)
	if err != nil {
		return nil, err
	}

	session.User = user
	sc.Sessions[sessionID] = session
	return session, nil
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
