package web

import (
	"crypto/hmac"
	"database/sql"
	"net/url"
	"time"

	logger "github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
)

const (

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2
)

const (
	// LenSessionID is the byte length of a session id.
	LenSessionID = 64
	// LenSessionIDBase64 is the length of a session id base64 encoded.
	LenSessionIDBase64 = 88
	// ErrSessionIDEmpty is thrown if a session id is empty.
	ErrSessionIDEmpty = Error("auth session id is empty")
	// ErrSessionIDTooLong is thrown if a session id is too long.
	ErrSessionIDTooLong = Error("auth session id is too long")

	// ErrSecureSessionIDEmpty is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDEmpty = Error("auth secure session id is empty")
	// ErrSecureSessionIDTooLong is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDTooLong = Error("auth secure session id is too long")
	// ErrSecureSessionIDInvalid is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDInvalid = Error("auth secure session id is invalid")
)

// IsErrSessionInvalid returns if an error is a session invalid error.
func IsErrSessionInvalid(err error) bool {
	if err == nil {
		return false
	}
	switch err {
	case ErrSessionIDEmpty,
		ErrSessionIDTooLong,
		ErrSecureSessionIDEmpty,
		ErrSecureSessionIDTooLong,
		ErrSecureSessionIDInvalid:
		return true
	default:
		return false
	}
}

// NewAuthManager returns a new session manager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		sessionCache:             NewSessionCache(),
		useSessionCache:          DefaultUseSessionCache,
		sessionTimeout:           DefaultSessionTimeout,
		sessionTimeoutIsAbsolute: DefaultSessionTimeoutIsAbsolute,
		cookieName:               DefaultCookieName,
		cookiePath:               DefaultCookiePath,
		secureCookieName:         DefaultSecureCookieName,
		secureCookiePath:         DefaultCookiePath,
	}
}

// NewAuthManagerFromConfig returns a new auth manager from a given config.
func NewAuthManagerFromConfig(cfg *AuthManagerConfig) *AuthManager {
	return &AuthManager{
		sessionCache:             NewSessionCache(),
		useSessionCache:          cfg.GetUseSessionCache(),
		sessionTimeout:           cfg.GetSessionTimeout(),
		sessionTimeoutIsAbsolute: cfg.GetSessionTimeoutIsAbsolute(),
		cookieHTTPS:              cfg.GetCookieHTTPS(),
		cookieName:               cfg.GetCookieName(),
		cookiePath:               cfg.GetCookiePath(),
		secureCookieKey:          cfg.GetSecureCookieKey(),
		secureCookieHTTPS:        cfg.GetSecureCookieHTTPS(),
		secureCookieName:         cfg.GetSecureCookieName(),
		secureCookiePath:         cfg.GetSecureCookiePath(),
	}
}

// AuthManager is a manager for sessions.
type AuthManager struct {
	useSessionCache      bool
	sessionCache         *SessionCache
	persistHandler       func(*Ctx, *Session, *sql.Tx) error
	fetchHandler         func(sessionID string, tx *sql.Tx) (*Session, error)
	removeHandler        func(sessionID string, tx *sql.Tx) error
	validateHandler      func(*Session, *sql.Tx) error
	loginRedirectHandler func(*url.URL) *url.URL

	log *logger.Logger

	sessionTimeout           time.Duration
	sessionTimeoutIsAbsolute bool
	sessionTimeoutProvider   func(rc *Ctx) *time.Time

	cookieName  string
	cookiePath  string
	cookieHTTPS bool

	secureCookieHTTPS bool
	secureCookieKey   []byte
	secureCookieName  string
	secureCookiePath  string
}

// --------------------------------------------------------------------------------
// Methods
// --------------------------------------------------------------------------------

// Login logs a userID in.
func (am *AuthManager) Login(userID string, ctx *Ctx) (session *Session, err error) {
	var sessionID string
	var secureSessionID string

	sessionID = am.createSessionID()
	if am.shouldIssueSecureSesssionID() {
		secureSessionID, err = am.createSecureSessionID(sessionID)
		if err != nil {
			return nil, err
		}
	}

	session = NewSession(userID, sessionID)
	session.ExpiresUTC = am.GenerateSessionTimeout(ctx)

	if am.persistHandler != nil {
		err = am.persistHandler(ctx, session, Tx(ctx))
		if err != nil {
			return nil, am.err(err)
		}
	}

	if am.useSessionCache {
		am.sessionCache.Upsert(session)
	}

	am.injectCookie(ctx, sessionID, session.ExpiresUTC)
	if am.shouldIssueSecureSesssionID() {
		am.injectSecureCookie(ctx, secureSessionID, session.ExpiresUTC)
	}
	return session, nil
}

// Logout unauthenticates a session.
func (am *AuthManager) Logout(session *Session, ctx *Ctx) error {
	if session == nil {
		return nil
	}

	// remove from session cache if enabled
	if am.useSessionCache {
		am.sessionCache.Remove(session.SessionID)
	}

	// expire cookies on the request
	if ctx != nil {
		ctx.ExpireCookie(am.cookieName, am.cookiePath)
		if am.shouldIssueSecureSesssionID() {
			ctx.ExpireCookie(am.secureCookieName, am.secureCookiePath)
		}
	}

	// remove the session from a backing store
	if am.removeHandler != nil {
		if ctx != nil {
			return am.err(am.removeHandler(session.SessionID, Tx(ctx)))
		}
		return am.err(am.removeHandler(session.SessionID, nil))
	}
	return nil
}

// VerifySession checks a sessionID to see if it's valid.
// It also handles updating a rolling expiry.
func (am *AuthManager) VerifySession(ctx *Ctx) (*Session, error) {
	sessionID := am.readSessionID(ctx)
	err := am.validateSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	var secureSessionID string
	if am.shouldIssueSecureSesssionID() {
		secureSessionID = am.readSecureSessionID(ctx)
		err := am.validateSecureSessionID(sessionID, secureSessionID)
		if err != nil {
			return nil, err
		}
	}

	var session *Session
	if am.useSessionCache {
		session = am.sessionCache.Get(sessionID)
	}

	if session == nil && am.fetchHandler != nil {
		session, err = am.fetchHandler(sessionID, Tx(ctx))
		if err != nil {
			return nil, err
		}
	}

	if session == nil || session.IsZero() || session.IsExpired() {
		if ctx != nil {
			ctx.ExpireCookie(am.cookieName, DefaultCookiePath)
			if am.shouldIssueSecureSesssionID() {
				ctx.ExpireCookie(am.cookieName, DefaultCookiePath)
			}
		}
		// if we have a remove handler and the sessionID is set
		if am.removeHandler != nil && len(sessionID) > 0 {
			err = am.removeHandler(sessionID, Tx(ctx))
			if err != nil {
				return nil, err
			}
		}
		// exit out, the session is bad
		return nil, nil
	}

	if am.validateHandler != nil {
		err = am.validateHandler(session, Tx(ctx))
		if err != nil {
			return nil, err
		}
	}

	// check if we need to do a rolling expiry update
	// note this will be explicitly false by default
	// as we use absolte expiry by default.
	if am.shouldUpdateSessionExpiry() {
		session.ExpiresUTC = am.GenerateSessionTimeout(ctx)
		if am.persistHandler != nil {
			err = am.persistHandler(ctx, session, Tx(ctx))
			if err != nil {
				return nil, err
			}
		}
		if ctx != nil {
			am.injectCookie(ctx, sessionID, session.ExpiresUTC)
			if am.shouldIssueSecureSesssionID() {
				am.injectSecureCookie(ctx, secureSessionID, session.ExpiresUTC)
			}
		}
	}

	if am.useSessionCache {
		am.sessionCache.Upsert(session)
	}
	return session, nil
}

// Redirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (am *AuthManager) Redirect(context *Ctx) Result {
	if am.loginRedirectHandler != nil {
		redirectTo := context.auth.loginRedirectHandler(context.Request.URL)
		if redirectTo != nil {
			return context.Redirectf(redirectTo.String())
		}
	}
	return context.DefaultResultProvider().NotAuthorized()
}

// --------------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------------

// SetSecret sets the secret for the auth manager.
func (am *AuthManager) SetSecret(secret []byte) {
	am.secureCookieKey = secret
}

// Secret returns the auth manager secret.
func (am *AuthManager) Secret() []byte {
	return am.secureCookieKey
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (am *AuthManager) SetCookiesAsSessionBound() {
	am.sessionTimeout = 0
	am.sessionTimeoutProvider = nil
}

// SetSessionTimeout sets the static value for session timeout.
func (am *AuthManager) SetSessionTimeout(timeout time.Duration) {
	am.sessionTimeout = timeout
}

// SetSessionTimeoutIsAbsolute sets if the timeout for session should be an absolute (vs. relative) time.
func (am *AuthManager) SetSessionTimeoutIsAbsolute(isAbsolute bool) {
	am.sessionTimeoutIsAbsolute = isAbsolute
}

// SetSessionTimeoutProvider sets the session to expire with a given the given timeout provider.
func (am *AuthManager) SetSessionTimeoutProvider(timeoutProvider func(rc *Ctx) *time.Time) {
	am.sessionTimeoutProvider = timeoutProvider
}

// SetCookieHTTPS overrides defaults when determining if we should use the HTTPS only cooikie option.
// The default depends on the app configuration (if tls is configured and enabled).
func (am *AuthManager) SetCookieHTTPS(isHTTPSOnly bool) {
	am.cookieHTTPS = isHTTPSOnly
}

// CookieHTTPS returns if the cookie is for only https connections.
func (am *AuthManager) CookieHTTPS() bool {
	return am.cookieHTTPS
}

// SetCookieName sets the session cookie name.
func (am *AuthManager) SetCookieName(paramName string) {
	am.cookieName = paramName
}

// CookieName returns the session param name.
func (am *AuthManager) CookieName() string {
	return am.cookieName
}

// SetCookiePath sets the session cookie path.
func (am *AuthManager) SetCookiePath(path string) {
	am.cookiePath = path
}

// CookiePath returns the session param path.
func (am *AuthManager) CookiePath() string {
	if len(am.cookiePath) == 0 {
		return DefaultCookiePath
	}
	return am.cookiePath
}

// SetSecureCookieName sets the session param name.
func (am *AuthManager) SetSecureCookieName(paramName string) {
	am.secureCookieName = paramName
}

// SecureCookieName returns the session param name.
func (am *AuthManager) SecureCookieName() string {
	return am.secureCookieName
}

// SecureCookiePath returns the secure session path.
func (am *AuthManager) SecureCookiePath() string {
	if len(am.secureCookiePath) > 0 {
		return am.secureCookiePath
	}
	if len(am.cookiePath) > 0 {
		return am.cookiePath
	}
	return DefaultCookiePath
}

// SetSecureCookiePath sets the secure session param path.
func (am *AuthManager) SetSecureCookiePath(path string) {
	am.secureCookiePath = path
}

// SetPersistHandler sets the persist handler.
// It must be able to both create sessions and update sessions if the expiry changes.
func (am *AuthManager) SetPersistHandler(handler func(*Ctx, *Session, *sql.Tx) error) {
	am.persistHandler = handler
}

// SetFetchHandler sets the fetch handler.
// It should return a session by a string sessionID.
func (am *AuthManager) SetFetchHandler(handler func(sessionID string, tx *sql.Tx) (*Session, error)) {
	am.fetchHandler = handler
}

// SetRemoveHandler sets the remove handler.
// It should remove a session from the backing store by a string sessionID.
func (am *AuthManager) SetRemoveHandler(handler func(sessionID string, tx *sql.Tx) error) {
	am.removeHandler = handler
}

// SetValidateHandler sets the validate handler.
// This is an optional handler that will evaluate the session when verifying requests that are session aware.
func (am *AuthManager) SetValidateHandler(handler func(*Session, *sql.Tx) error) {
	am.validateHandler = handler
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template, provided one is configured.
func (am *AuthManager) SetLoginRedirectHandler(handler func(*url.URL) *url.URL) {
	am.loginRedirectHandler = handler
}

// SessionCache returns the session cache.
func (am *AuthManager) SessionCache() *SessionCache {
	return am.sessionCache
}

// IsCookieHTTPSOnly returns if the session cookie is configured to be secure only.
func (am *AuthManager) IsCookieHTTPSOnly() bool {
	return am.cookieHTTPS
}

// IsSecureCookieHTTPSOnly returns if the secure session cookie is configured to be secure only.
func (am *AuthManager) IsSecureCookieHTTPSOnly() bool {
	return am.secureCookieHTTPS
}

// GenerateSessionTimeout returns the absolute time the cookie would expire.
func (am *AuthManager) GenerateSessionTimeout(context *Ctx) *time.Time {
	if am.sessionTimeout > 0 {
		return util.OptionalTime(time.Now().UTC().Add(am.sessionTimeout))
	} else if am.sessionTimeoutProvider != nil {
		return am.sessionTimeoutProvider(context)
	}
	return nil
}

// SetLogger sets the intance logger.
func (am *AuthManager) SetLogger(log *logger.Logger) {
	am.log = log
}

// WithLogger sets the intance logger and returns a reference.
func (am *AuthManager) WithLogger(log *logger.Logger) *AuthManager {
	am.log = log
	return am
}

// Logger returns the instance logger.
func (am *AuthManager) Logger() *logger.Logger {
	return am.log
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

func (am *AuthManager) shouldIssueSecureSesssionID() bool {
	return len(am.secureCookieKey) > 0
}

func (am AuthManager) shouldUpdateSessionExpiry() bool {
	return !am.sessionTimeoutIsAbsolute && (am.sessionTimeout > 0 || am.sessionTimeoutProvider != nil)
}

// CreateSessionID creates a new session id.
func (am AuthManager) createSessionID() string {
	return NewSessionID()
}

// CreateSecureSessionID creates a secure session id.
func (am AuthManager) createSecureSessionID(sessionID string) (string, error) {
	return EncodeSignSessionID(sessionID, am.secureCookieKey)
}

// InjectCookie injects a session cookie into the context.
func (am *AuthManager) injectCookie(ctx *Ctx, sessionID string, expire *time.Time) {
	paramName := am.CookieName()
	path := am.CookiePath()
	https := am.CookieHTTPS()
	if ctx != nil {
		ctx.WriteNewCookie(paramName, sessionID, expire, path, https)
	}
}

// InjectCookie injects a session cookie into the context.
func (am *AuthManager) injectSecureCookie(ctx *Ctx, sessionID string, expire *time.Time) {
	paramName := am.SecureCookieName()
	path := am.SecureCookiePath()
	https := am.IsSecureCookieHTTPSOnly()
	if ctx != nil {
		ctx.WriteNewCookie(paramName, sessionID, expire, path, https)
	}
}

// ReadSessionID reads a session id from a given request context.
func (am *AuthManager) readSessionID(context *Ctx) string {
	if cookie := context.GetCookie(am.CookieName()); cookie != nil {
		return cookie.Value
	}

	if headerValue, err := context.HeaderParam(am.CookieName()); err == nil {
		return headerValue
	}

	return ""
}

// ReadSecureSessionID reads a secure session id from a given request context.
func (am *AuthManager) readSecureSessionID(context *Ctx) string {
	if cookie := context.GetCookie(am.SecureCookieName()); cookie != nil {
		return cookie.Value
	}

	if headerValue, err := context.HeaderParam(am.SecureCookieName()); err == nil {
		return headerValue
	}

	return ""
}

// ValidateSessionID verifies a session id.
func (am *AuthManager) validateSessionID(sessionID string) error {
	if len(sessionID) == 0 {
		return ErrSessionIDEmpty
	}
	if len(sessionID) > LenSessionIDBase64 {
		return ErrSessionIDTooLong
	}
	return nil
}

// ValidateSecureSessionID verifies a session id.
func (am *AuthManager) validateSecureSessionID(sessionID, secureSessionID string) error {
	if len(secureSessionID) == 0 {
		return ErrSecureSessionIDEmpty
	}

	if len(secureSessionID) > LenSessionIDBase64 {
		return ErrSecureSessionIDTooLong
	}

	secureSessionIDDecoded, err := Base64.Decode(secureSessionID)
	if err != nil {
		return ErrSecureSessionIDInvalid
	}

	signedSessionID, err := SignSessionID(sessionID, am.secureCookieKey)
	if err != nil {
		return ErrSecureSessionIDInvalid
	}
	if !hmac.Equal(signedSessionID, secureSessionIDDecoded) {
		return ErrSecureSessionIDInvalid
	}

	return nil
}

func (am AuthManager) debugf(format string, args ...interface{}) {
	if am.log != nil {
		am.log.SyncDebugf(format, args...)
	}
}

func (am AuthManager) err(err error) error {
	if am.log != nil {
		am.log.Error(err)
	}
	return err
}
