package web

import (
	"database/sql"
	"net/url"
	"time"

	exception "github.com/blendlabs/go-exception"
)

const (
	// SessionParamName is the name of the field (header, cookie, or querystring) that needs to have the sessionID on it.
	SessionParamName = "__auth_token__"

	// StateKeySession is the state key for the user session.
	StateKeySession = "__session__"

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2
)

// NewSessionManager returns a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessionCache:           NewSessionCache(),
		sessionCookieIsSession: true,
	}
}

// SessionManager is a manager for sessions.
type SessionManager struct {
	sessionCache         *SessionCache
	persistHandler       func(*Session, *sql.Tx) error
	fetchHandler         func(sessionID string, tx *sql.Tx) (*Session, error)
	removeHandler        func(sessionID string, tx *sql.Tx) error
	validateHandler      func(*Session, *sql.Tx) error
	loginRedirectHandler func(*url.URL) *url.URL

	sessionCookieIsSession bool
	sessionCookieTimeout   func() *time.Time
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (sm *SessionManager) SetCookiesAsSessionBound() {
	sm.sessionCookieIsSession = true
	sm.sessionCookieTimeout = nil
}

// SetCookieTimeout sets the cookies to the given timeout.
func (sm *SessionManager) SetCookieTimeout(timeoutFunc func() *time.Time) {
	sm.sessionCookieIsSession = false
	sm.sessionCookieTimeout = timeoutFunc
}

// CreateSessionID creates a new session id.
func (sm SessionManager) CreateSessionID() string {
	return String.SecureRandom(64)
}

// SetPersistHandler sets the persist handler
func (sm *SessionManager) SetPersistHandler(handler func(*Session, *sql.Tx) error) {
	sm.persistHandler = handler
}

// SetFetchHandler sets the fetch handler
func (sm *SessionManager) SetFetchHandler(handler func(sessionID string, tx *sql.Tx) (*Session, error)) {
	sm.fetchHandler = handler
}

// SetRemoveHandler sets the remove handler.
func (sm *SessionManager) SetRemoveHandler(handler func(sessionID string, tx *sql.Tx) error) {
	sm.removeHandler = handler
}

// SetValidateHandler sets the validate handler.
func (sm *SessionManager) SetValidateHandler(handler func(*Session, *sql.Tx) error) {
	sm.validateHandler = handler
}

// InjectSession injects the session object into a request context.
func (sm SessionManager) InjectSession(session *Session, context *RequestContext) {
	context.SetState(StateKeySession, session)
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template.
func (sm *SessionManager) SetLoginRedirectHandler(handler func(*url.URL) *url.URL) {
	sm.loginRedirectHandler = handler
}

// SessionCache returns the session cache.
func (sm SessionManager) SessionCache() *SessionCache {
	return sm.sessionCache
}

// GetSession extracts the session from the RequestContext
func (sm SessionManager) GetSession(context *RequestContext) *Session {
	if sessionStorage := context.State(StateKeySession); sessionStorage != nil {
		if session, isSession := sessionStorage.(*Session); isSession {
			return session
		}
	}
	return nil
}

// Login logs a userID in.
func (sm *SessionManager) Login(userID int64, context *RequestContext) (*Session, error) {
	sessionID := sm.CreateSessionID()
	session := NewSession(userID, sessionID)

	var err error
	if sm.persistHandler != nil {
		err = sm.persistHandler(session, context.Tx())
		if err != nil {
			return nil, err
		}
	}

	sm.sessionCache.Add(session)
	if context != nil {
		if sm.sessionCookieIsSession {
			context.SetCookie(SessionParamName, sessionID, nil, "/")
		} else if sm.sessionCookieTimeout != nil {
			context.SetCookie(SessionParamName, sessionID, sm.sessionCookieTimeout(), "/")
		}
	}
	return session, nil
}

// Logout un-authenticates a session.
func (sm *SessionManager) Logout(userID int64, sessionID string, context *RequestContext) error {
	sm.sessionCache.Expire(sessionID)

	if context != nil {
		context.ExpireCookie(SessionParamName)
	}
	if sm.removeHandler != nil {
		if context != nil {
			return sm.removeHandler(sessionID, context.Tx())
		}
		return sm.removeHandler(sessionID, nil)
	}
	return nil
}

// VerifySession checks a sessionID to see if it's valid.
func (sm *SessionManager) VerifySession(sessionID string, context *RequestContext) (*Session, error) {
	if sm.sessionCache.IsActive(sessionID) {
		return sm.sessionCache.Get(sessionID), nil
	}

	if sm.fetchHandler == nil {
		return nil, exception.New("Must provide a `fetchHandler` to retrieve dormant sessions.")
	}

	var session *Session
	var err error
	if context != nil {
		session, err = sm.fetchHandler(sessionID, context.Tx())
	} else {
		session, err = sm.fetchHandler(sessionID, nil)
	}
	if err != nil {
		return nil, err
	}
	if session != nil {
		sm.sessionCache.Add(session)
	}
	return session, nil
}
