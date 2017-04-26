package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"io"
	"net/url"
	"time"
)

const (
	// DefaultSessionParamName is the default name of the field that contains the session id.
	DefaultSessionParamName = "SID"

	// DefaultSecureSessionParamName is the default name of the field that contains the secure session id.
	DefaultSecureSessionParamName = "SSID"

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2

	// DefaultSessionCookiePath is the default cookie path.
	DefaultSessionCookiePath = "/"
)

const (
	// ErrSessionIDInvalid is an error that is thrown if a given session id is invalid.
	ErrSessionIDInvalid = Error("provided session id is invalid")
)

// NewAuthManager returns a new session manager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		sessionCache:                NewSessionCache(),
		sessionCookieIsSessionBound: true,
		sessionParamName:            DefaultSessionParamName,
		secureSessionParamName:      DefaultSecureSessionParamName,
	}
}

// AuthManager is a manager for sessions.
type AuthManager struct {
	sessionCache           *SessionCache
	persistHandler         func(*Ctx, *Session, *sql.Tx) error
	fetchHandler           func(sessionID string, tx *sql.Tx) (*Session, error)
	removeHandler          func(sessionID string, tx *sql.Tx) error
	validateHandler        func(*Session, *sql.Tx) error
	loginRedirectHandler   func(*url.URL) *url.URL
	sessionParamName       string
	secureSessionParamName string

	secret []byte

	sessionCookieIsSessionBound  bool
	sessionCookieIsHTTPSOnly     bool
	sessionCookieTimeoutProvider func(rc *Ctx) *time.Time
}

// --------------------------------------------------------------------------------
// Methods
// --------------------------------------------------------------------------------

// Login logs a userID in.
func (am *AuthManager) Login(userID int64, context *Ctx) (session *Session, err error) {
	var sessionID string
	var secureSessionID string

	sessionID = am.createSessionID()

	if am.ShouldIssueSecureSesssionID() {
		secureSessionID, err = am.createSecureSessionID(sessionID)
		if err != nil {
			return nil, err
		}
	}

	session = NewSession(userID, sessionID)
	if am.persistHandler != nil {
		err = am.persistHandler(context, session, context.Tx())
		if err != nil {
			return nil, err
		}
	}

	am.sessionCache.Add(session)
	am.injectCookie(am.sessionParamName, context, sessionID)
	if am.ShouldIssueSecureSesssionID() {
		am.injectCookie(am.secureSessionParamName, context, secureSessionID)
	}
	return session, nil
}

// Logout unauthenticates a session.
func (am *AuthManager) Logout(session *Session, context *Ctx) error {
	if session == nil {
		return nil
	}

	am.sessionCache.Expire(session.SessionID)

	if context != nil {
		context.ExpireCookie(am.sessionParamName, DefaultSessionCookiePath)
		if am.ShouldIssueSecureSesssionID() {
			context.ExpireCookie(am.secureSessionParamName, DefaultSessionCookiePath)
		}
	}
	if am.removeHandler != nil {
		if context != nil {
			return am.removeHandler(session.SessionID, context.Tx())
		}
		return am.removeHandler(session.SessionID, nil)
	}
	return nil
}

// VerifySession checks a sessionID to see if it's valid.
func (am *AuthManager) VerifySession(context *Ctx) (*Session, error) {
	sessionID := am.readSessionID(context)

	err := am.validateSessionID(sessionID)
	if err != nil {
		return nil, nil
	}

	if am.ShouldIssueSecureSesssionID() {
		secureSessionID := am.readSecureSessionID(context)
		err := am.validateSecureSessionID(sessionID, secureSessionID)
		if err != nil {
			return nil, nil
		}
	}

	if session, hasSession := am.sessionCache.Get(sessionID); hasSession {
		return session, nil
	}

	if am.fetchHandler == nil {
		if context != nil {
			context.ExpireCookie(am.sessionParamName, DefaultSessionCookiePath)
		}
		return nil, nil
	}

	var session *Session
	if context != nil {
		session, err = am.fetchHandler(sessionID, context.Tx())
	} else {
		session, err = am.fetchHandler(sessionID, nil)
	}
	if err != nil {
		return nil, err
	}
	if session == nil || session.IsZero() {
		if context != nil {
			context.ExpireCookie(am.sessionParamName, DefaultSessionCookiePath)
			if am.ShouldIssueSecureSesssionID() {
				context.ExpireCookie(am.secureSessionParamName, DefaultSessionCookiePath)
			}
		}
		return nil, nil
	}

	if am.validateHandler != nil {
		if context != nil {
			err = am.validateHandler(session, context.Tx())
		} else {
			err = am.validateHandler(session, context.Tx())
		}
		if err != nil {
			return nil, err
		}
	}

	am.sessionCache.Add(session)
	return session, nil
}

// Redirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (am *AuthManager) Redirect(context *Ctx) Result {
	if am.loginRedirectHandler != nil {
		redirectTo := context.auth.loginRedirectHandler(context.Request.URL)
		if redirectTo != nil {
			return context.Redirect(redirectTo.String())
		}
	}
	return context.DefaultResultProvider().NotAuthorized()
}

// --------------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------------

// SetSecret sets the secret for the auth manager.
func (am *AuthManager) SetSecret(secret []byte) {
	am.secret = secret
}

// Secret returns the auth manager secret.
func (am *AuthManager) Secret() []byte {
	return am.secret
}

// ShouldIssueSecureSesssionID indicates if we shoul issue a second secure sessionID to check the sessionID.
func (am *AuthManager) ShouldIssueSecureSesssionID() bool {
	return len(am.secret) > 0
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (am *AuthManager) SetCookiesAsSessionBound() {
	am.sessionCookieIsSessionBound = true
	am.sessionCookieTimeoutProvider = nil
}

// SetCookieTimeout sets the cookies to the given timeout.
func (am *AuthManager) SetCookieTimeout(timeoutProvider func(rc *Ctx) *time.Time) {
	am.sessionCookieIsSessionBound = false
	am.sessionCookieTimeoutProvider = timeoutProvider
}

// SetCookieAsHTTPSOnly overrides defaults when determining if we should use the HTTPS only cooikie option.
// The default depends on the app configuration (if tls is configured and enabled).
func (am *AuthManager) SetCookieAsHTTPSOnly(isHTTPSOnly bool) {
	am.sessionCookieIsHTTPSOnly = isHTTPSOnly
}

// SetSessionParamName sets the session param name.
func (am *AuthManager) SetSessionParamName(paramName string) {
	am.sessionParamName = paramName
}

// SecureSessionParamName returns the session param name.
func (am *AuthManager) SecureSessionParamName() string {
	return am.secureSessionParamName
}

// SetSecureSessionParamName sets the session param name.
func (am *AuthManager) SetSecureSessionParamName(paramName string) {
	am.secureSessionParamName = paramName
}

// SessionParamName returns the session param name.
func (am *AuthManager) SessionParamName() string {
	return am.sessionParamName
}

// SetPersistHandler sets the persist handler
func (am *AuthManager) SetPersistHandler(handler func(*Ctx, *Session, *sql.Tx) error) {
	am.persistHandler = handler
}

// SetFetchHandler sets the fetch handler
func (am *AuthManager) SetFetchHandler(handler func(sessionID string, tx *sql.Tx) (*Session, error)) {
	am.fetchHandler = handler
}

// SetRemoveHandler sets the remove handler.
func (am *AuthManager) SetRemoveHandler(handler func(sessionID string, tx *sql.Tx) error) {
	am.removeHandler = handler
}

// SetValidateHandler sets the validate handler.
func (am *AuthManager) SetValidateHandler(handler func(*Session, *sql.Tx) error) {
	am.validateHandler = handler
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template.
func (am *AuthManager) SetLoginRedirectHandler(handler func(*url.URL) *url.URL) {
	am.loginRedirectHandler = handler
}

// SessionCache returns the session cache.
func (am AuthManager) SessionCache() *SessionCache {
	return am.sessionCache
}

// IsCookieHTTPSOnly returns if the session cookie is configured to be secure only.
func (am *AuthManager) IsCookieHTTPSOnly() bool {
	return am.sessionCookieIsHTTPSOnly
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// CreateSessionID creates a new session id.
func (am AuthManager) createSessionID() string {
	return NewSessionID()
}

// CreateSecureSessionID creates a secure session id.
func (am AuthManager) createSecureSessionID(sessionID string) (string, error) {
	return EncodeSignSessionID(sessionID, am.secret)
}

// InjectCookie injects a session cookie into the context.
func (am *AuthManager) injectCookie(paramName string, context *Ctx, sessionID string) {
	if context != nil {
		if am.sessionCookieIsSessionBound {
			context.WriteNewCookie(paramName, sessionID, nil, DefaultSessionCookiePath, am.IsCookieHTTPSOnly())
		} else if am.sessionCookieTimeoutProvider != nil {
			context.WriteNewCookie(paramName, sessionID, am.sessionCookieTimeoutProvider(context), DefaultSessionCookiePath, am.IsCookieHTTPSOnly())
		}
	}
}

// ReadSessionID reads a session id from a given request context.
func (am *AuthManager) readSessionID(context *Ctx) string {
	if cookie := context.GetCookie(am.SessionParamName()); cookie != nil {
		return cookie.Value
	}

	if headerValue, err := context.HeaderParam(am.SessionParamName()); err == nil {
		return headerValue
	}

	return ""
}

// ReadSecureSessionID reads a secure session id from a given request context.
func (am *AuthManager) readSecureSessionID(context *Ctx) string {
	if cookie := context.GetCookie(am.SecureSessionParamName()); cookie != nil {
		return cookie.Value
	}

	if headerValue, err := context.HeaderParam(am.SecureSessionParamName()); err == nil {
		return headerValue
	}

	return ""
}

// ValidateSessionID verifies a session id.
func (am *AuthManager) validateSessionID(sessionID string) error {
	if len(sessionID) == 0 || len(sessionID) > 128 {
		return ErrSessionIDInvalid
	}
	return nil
}

// ValidateSecureSessionID verifies a session id.
func (am *AuthManager) validateSecureSessionID(sessionID, secureSessionID string) error {
	if len(secureSessionID) == 0 || len(secureSessionID) > 128 {
		return ErrSessionIDInvalid
	}

	secureSessionIDDecoded, err := Base64.Decode(secureSessionID)
	if err != nil {
		return ErrSessionIDInvalid
	}

	signedSessionID, err := SignSessionID(sessionID, am.secret)
	if err != nil {
		return ErrSessionIDInvalid
	}
	if !hmac.Equal(signedSessionID, secureSessionIDDecoded) {
		return ErrSessionIDInvalid
	}

	return nil
}

// --------------------------------------------------------------------------------
// Utility Functions
// --------------------------------------------------------------------------------

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	return String.SecureRandom(64)
}

// SignSessionID returns a new secure session id.
func SignSessionID(sessionID string, key []byte) ([]byte, error) {
	mac := hmac.New(sha512.New, key)
	_, err := mac.Write([]byte(sessionID))
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}

// EncodeSignSessionID returns a new secure session id base64 encoded..
func EncodeSignSessionID(sessionID string, key []byte) (string, error) {
	signed, err := SignSessionID(sessionID, key)
	if err != nil {
		return "", err
	}
	return Base64.Encode(signed), nil
}

// GenerateCryptoKey generates a cryptographic key.
func GenerateCryptoKey(keySize int) []byte {
	key := make([]byte, keySize)
	io.ReadFull(rand.Reader, key)
	return key
}

// GenerateSHA512Key generates a crypto key for SHA512 hashing.
func GenerateSHA512Key() []byte {
	return GenerateCryptoKey(64)
}
