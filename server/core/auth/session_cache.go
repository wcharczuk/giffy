package auth

import (
	"sync"
	"time"

	"github.com/wcharczuk/giffy/server/model"
)

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
func (sc *SessionCache) Add(userID int64, sessionID string) (*Session, error) {
	session := NewSession(userID, sessionID)

	user, err := model.GetUserByID(session.UserID, nil)
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

// NewSession returns a new session object.
func NewSession(userID int64, sessionID string) *Session {
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
		State:      map[string]interface{}{},
	}
}

// Session is an active session
type Session struct {
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	CreatedUTC time.Time              `json:"created_utc"`
	User       *model.User            `json:"user"`
	State      map[string]interface{} `json:"-"`
}

var (
	sessionCacheLatch = sync.Mutex{}
	sessionCache      *SessionCache
)

// SessionState returns the shared SessionCache singleton.
func SessionState() *SessionCache {
	sessionCacheLatch.Lock()
	defer sessionCacheLatch.Unlock()

	if sessionCache == nil {
		sessionCache = NewSessionCache()
	}

	return sessionCache
}

// LockSessionState locks the session state object for the caller.
func LockSessionState() {
	sessionCacheLatch.Lock()
}

// UnlockSessionState unlocks the session state object for the caller.
func UnlockSessionState() {
	sessionCacheLatch.Unlock()
}
