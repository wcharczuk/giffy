package web

import (
	"database/sql"
	"net/url"
	"time"
)

const (
	// DefaultSessionParamName is the default name of the field (header, cookie, or querystring) that needs to have the sessionID on it.
	DefaultSessionParamName = "__auth_token__"

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2

	// DefaultSessionCookiePath is the default cookie path.
	DefaultSessionCookiePath = "/"
)

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	return String.SecureRandom(64)
}

// NewAuthManager returns a new session manager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		sessionCache:                NewSessionCache(),
		sessionCookieIsSessionBound: true,
		sessionParamName:            DefaultSessionParamName,
	}
}

// AuthManager is a manager for sessions.
type AuthManager struct {
	sessionCache         *SessionCache
	persistHandler       func(*Ctx, *Session, *sql.Tx) error
	fetchHandler         func(sessionID string, tx *sql.Tx) (*Session, error)
	removeHandler        func(sessionID string, tx *sql.Tx) error
	validateHandler      func(*Session, *sql.Tx) error
	loginRedirectHandler func(*url.URL) *url.URL
	sessionParamName     string

	sessionCookieIsSessionBound  bool
	sessionCookieIsSecure        *bool
	sessionCookieTimeoutProvider func(rc *Ctx) *time.Time
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (sm *AuthManager) SetCookiesAsSessionBound() {
	sm.sessionCookieIsSessionBound = true
	sm.sessionCookieTimeoutProvider = nil
}

// SetCookieTimeout sets the cookies to the given timeout.
func (sm *AuthManager) SetCookieTimeout(timeoutProvider func(rc *Ctx) *time.Time) {
	sm.sessionCookieIsSessionBound = false
	sm.sessionCookieTimeoutProvider = timeoutProvider
}

// SetCookieAsSecure overrides defaults when determining if we should use the HTTPS only cooikie option.
// The default depends on the app configuration (if tls is configured and enabled).
func (sm *AuthManager) SetCookieAsSecure(isSecure bool) {
	sm.sessionCookieIsSecure = &isSecure
}

// SessionParamName returns the session param name.
func (sm *AuthManager) SessionParamName() string {
	return sm.sessionParamName
}

// SetSessionParamName sets the session param name.
func (sm *AuthManager) SetSessionParamName(paramName string) {
	sm.sessionParamName = paramName
}

// SetPersistHandler sets the persist handler
func (sm *AuthManager) SetPersistHandler(handler func(*Ctx, *Session, *sql.Tx) error) {
	sm.persistHandler = handler
}

// SetFetchHandler sets the fetch handler
func (sm *AuthManager) SetFetchHandler(handler func(sessionID string, tx *sql.Tx) (*Session, error)) {
	sm.fetchHandler = handler
}

// SetRemoveHandler sets the remove handler.
func (sm *AuthManager) SetRemoveHandler(handler func(sessionID string, tx *sql.Tx) error) {
	sm.removeHandler = handler
}

// SetValidateHandler sets the validate handler.
func (sm *AuthManager) SetValidateHandler(handler func(*Session, *sql.Tx) error) {
	sm.validateHandler = handler
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template.
func (sm *AuthManager) SetLoginRedirectHandler(handler func(*url.URL) *url.URL) {
	sm.loginRedirectHandler = handler
}

// SessionCache returns the session cache.
func (sm AuthManager) SessionCache() *SessionCache {
	return sm.sessionCache
}

// Login logs a userID in.
func (sm *AuthManager) Login(userID int64, context *Ctx) (*Session, error) {
	sessionID := NewSessionID()
	session := NewSession(userID, sessionID)

	var err error
	if sm.persistHandler != nil {
		err = sm.persistHandler(context, session, context.Tx())
		if err != nil {
			return nil, err
		}
	}

	sm.sessionCache.Add(session)
	sm.InjectSessionCookie(context, sessionID)
	return session, nil
}

// InjectSessionCookie injects a session cookie into the context.
func (sm *AuthManager) InjectSessionCookie(context *Ctx, sessionID string) {
	if context != nil {
		if sm.sessionCookieIsSessionBound {
			context.WriteNewCookie(sm.sessionParamName, sessionID, nil, DefaultSessionCookiePath, sm.IsCookieSecure())
		} else if sm.sessionCookieTimeoutProvider != nil {
			context.WriteNewCookie(sm.sessionParamName, sessionID, sm.sessionCookieTimeoutProvider(context), DefaultSessionCookiePath, sm.IsCookieSecure())
		}
	}
}

// IsCookieSecure returns if the session cookie is configured to be secure only.
func (sm *AuthManager) IsCookieSecure() bool {
	return sm.sessionCookieIsSecure != nil && *sm.sessionCookieIsSecure
}

// Logout un-authenticates a session.
func (sm *AuthManager) Logout(session *Session, context *Ctx) error {
	if session == nil {
		return nil
	}

	sm.sessionCache.Expire(session.SessionID)

	if context != nil {
		context.ExpireCookie(sm.sessionParamName, DefaultSessionCookiePath)
	}
	if sm.removeHandler != nil {
		if context != nil {
			return sm.removeHandler(session.SessionID, context.Tx())
		}
		return sm.removeHandler(session.SessionID, nil)
	}
	return nil
}

// ReadSessionID reads a session id from a given request context.
func (sm *AuthManager) ReadSessionID(context *Ctx) string {
	if headerValue, err := context.HeaderParam(sm.SessionParamName()); err == nil {
		return headerValue
	}

	if cookie := context.GetCookie(sm.SessionParamName()); cookie != nil {
		return cookie.Value
	}

	return ""
}

// VerifySession checks a sessionID to see if it's valid.
func (sm *AuthManager) VerifySession(sessionID string, context *Ctx) (*Session, error) {
	if sm.sessionCache.IsActive(sessionID) {
		return sm.sessionCache.Get(sessionID), nil
	}

	if sm.fetchHandler == nil {
		if context != nil {
			context.ExpireCookie(sm.sessionParamName, DefaultSessionCookiePath)
		}
		return nil, nil
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
	if session == nil || session.IsZero() {
		if context != nil {
			context.ExpireCookie(sm.sessionParamName, DefaultSessionCookiePath)
		}
		return nil, nil
	}

	sm.sessionCache.Add(session)
	return session, nil
}

// ReadAndVerifySession reads a session off a context and verifies it.
// Note this will also inject the session into the context.
func (sm *AuthManager) ReadAndVerifySession(context *Ctx) (*Session, error) {
	sessionID := sm.ReadSessionID(context)
	if len(sessionID) > 0 {
		session, err := sm.VerifySession(sessionID, context)
		if err != nil {
			return nil, err
		}

		return session, nil
	}

	return nil, nil
}

// Redirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (sm *AuthManager) Redirect(context *Ctx) Result {
	if sm.loginRedirectHandler != nil {
		redirectTo := context.auth.loginRedirectHandler(context.Request.URL)
		if redirectTo != nil {
			return context.Redirect(redirectTo.String())
		}
	}
	return context.DefaultResultProvider().NotAuthorized()
}
