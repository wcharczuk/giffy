package web

import (
	"crypto/hmac"
	"net/url"
	"time"

	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/util"
)

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
	}
}

// NewAuthManagerFromConfig returns a new auth manager from a given config.
func NewAuthManagerFromConfig(cfg *Config) *AuthManager {
	return &AuthManager{
		sessionCache:             NewSessionCache(),
		useSessionCache:          cfg.GetUseSessionCache(),
		sessionTimeout:           cfg.GetSessionTimeout(),
		sessionTimeoutIsAbsolute: cfg.GetSessionTimeoutIsAbsolute(),
		cookieHTTPSOnly:          cfg.GetCookieHTTPSOnly(),
		cookieName:               cfg.GetCookieName(),
		cookiePath:               cfg.GetCookiePath(),
		secret:                   cfg.GetAuthSecret(),
		secureCookieHTTPSOnly:    cfg.GetSecureCookieHTTPSOnly(),
		secureCookieName:         cfg.GetSecureCookieName(),
	}
}

// AuthManager is a manager for sessions.
type AuthManager struct {
	useSessionCache      bool
	sessionCache         *SessionCache
	persistHandler       func(*Ctx, *Session, State) error
	fetchHandler         func(sessionID string, state State) (*Session, error)
	removeHandler        func(sessionID string, state State) error
	validateHandler      func(*Session, State) error
	loginRedirectHandler func(*Ctx) *url.URL

	log *logger.Logger

	sessionTimeout           time.Duration
	sessionTimeoutIsAbsolute bool
	sessionTimeoutProvider   func(rc *Ctx) *time.Time

	cookieName      string
	cookiePath      string
	cookieHTTPSOnly bool

	secret                []byte
	secureCookieName      string
	secureCookieHTTPSOnly bool
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
		err = am.persistHandler(ctx, session, ctx.state)
		if err != nil {
			return nil, am.err(err)
		}
	}

	if am.useSessionCache {
		am.sessionCache.Upsert(session)
	}

	am.injectCookie(ctx, am.CookieName(), sessionID, session.ExpiresUTC)
	if am.shouldIssueSecureSesssionID() {
		am.injectCookie(ctx, am.SecureCookieName(), secureSessionID, session.ExpiresUTC)
	}
	return session, nil
}

// Logout unauthenticates a session.
func (am *AuthManager) Logout(ctx *Ctx) error {
	sessionID := am.readSessionID(ctx)

	// remove from session cache if enabled
	if am.useSessionCache {
		am.sessionCache.Remove(sessionID)
	}

	ctx.ExpireCookie(am.CookieName(), am.CookiePath())
	if am.shouldIssueSecureSesssionID() {
		ctx.ExpireCookie(am.SecureCookieName(), am.CookiePath())
	}
	ctx.WithSession(nil)

	// remove the session from a backing store
	if am.removeHandler != nil {
		return am.err(am.removeHandler(sessionID, ctx.state))
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
		session, err = am.fetchHandler(sessionID, ctx.state)
		if err != nil {
			return nil, err
		}
	}

	if session == nil || session.IsZero() || session.IsExpired() {
		ctx.ExpireCookie(am.CookieName(), DefaultCookiePath)
		if am.shouldIssueSecureSesssionID() {
			ctx.ExpireCookie(am.SecureCookieName(), am.CookiePath())
		}

		// if we have a remove handler and the sessionID is set
		if am.removeHandler != nil {
			err = am.removeHandler(sessionID, ctx.state)
			if err != nil {
				return nil, err
			}
		}

		// exit out, the session is bad
		return nil, nil
	}

	if am.validateHandler != nil {
		err = am.validateHandler(session, ctx.state)
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
			err = am.persistHandler(ctx, session, ctx.state)
			if err != nil {
				return nil, err
			}
		}

		am.injectCookie(ctx, am.CookieName(), sessionID, session.ExpiresUTC)
		if am.shouldIssueSecureSesssionID() {
			am.injectCookie(ctx, am.SecureCookieName(), secureSessionID, session.ExpiresUTC)
		}
	}

	if am.useSessionCache {
		am.sessionCache.Upsert(session)
	}
	return session, nil
}

// Redirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (am *AuthManager) Redirect(ctx *Ctx) Result {
	if am.loginRedirectHandler != nil {
		redirectTo := am.loginRedirectHandler(ctx)
		if redirectTo != nil {
			return ctx.Redirectf(redirectTo.String())
		}
	}
	return ctx.DefaultResultProvider().NotAuthorized()
}

// --------------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------------

// WithUseSessionCache sets if we should use the session cache.
func (am *AuthManager) WithUseSessionCache(value bool) *AuthManager {
	am.SetUseSessionCache(value)
	return am
}

// SetUseSessionCache sets the `UseSessionCache` property to the value.
func (am *AuthManager) SetUseSessionCache(value bool) {
	am.useSessionCache = value
}

// UseSessionCache returns if we should use the session cache.
func (am *AuthManager) UseSessionCache() bool {
	return am.useSessionCache
}

// WithSecret sets the secret for the auth manager.
func (am *AuthManager) WithSecret(secret []byte) *AuthManager {
	am.SetSecret(secret)
	return am
}

// SetSecret sets the secret for the auth manager.
func (am *AuthManager) SetSecret(secret []byte) {
	am.secret = secret
}

// Secret returns the auth manager secret.
func (am *AuthManager) Secret() []byte {
	return am.secret
}

// WithCookiesAsSessionBound sets cookies to be issued with `session` liveness.
func (am *AuthManager) WithCookiesAsSessionBound() *AuthManager {
	am.SetCookiesAsSessionBound()
	return am
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (am *AuthManager) SetCookiesAsSessionBound() {
	am.sessionTimeout = 0
	am.sessionTimeoutProvider = nil
}

// CookiesAsSessionBound returns if cookies are issued with `session` liveness.
func (am *AuthManager) CookiesAsSessionBound() bool {
	return am.sessionTimeout == 0 && am.sessionTimeoutProvider == nil
}

// WithSessionTimeout sets the either rolling or absolute session timeout.
func (am *AuthManager) WithSessionTimeout(timeout time.Duration) *AuthManager {
	am.SetSessionTimeout(timeout)
	return am
}

// SetSessionTimeout sets the static value for session timeout.
func (am *AuthManager) SetSessionTimeout(timeout time.Duration) {
	am.sessionTimeout = timeout
}

// SessionTimeout returns the session timeout.
func (am *AuthManager) SessionTimeout() time.Duration {
	return am.sessionTimeout
}

// WithAbsoluteSessionTimeout sets if the session timeout is absolute (vs. rolling).
func (am *AuthManager) WithAbsoluteSessionTimeout() *AuthManager {
	am.SetSessionTimeoutIsAbsolute(true)
	return am
}

// WithRollingSessionTimeout sets if the session timeout to be rolling (i.e. rolling).
func (am *AuthManager) WithRollingSessionTimeout() *AuthManager {
	am.SetSessionTimeoutIsAbsolute(false)
	return am
}

// SetSessionTimeoutIsAbsolute sets if the timeout for session should be an absolute (vs. rolling) time.
func (am *AuthManager) SetSessionTimeoutIsAbsolute(isAbsolute bool) {
	am.sessionTimeoutIsAbsolute = isAbsolute
}

// SesssionTimeoutIsAbsolute returns if the session timeout is absolute (vs. rolling).
func (am *AuthManager) SesssionTimeoutIsAbsolute() bool {
	return am.sessionTimeoutIsAbsolute
}

// SesssionTimeoutIsRolling returns if the session timeout is absolute (vs. rolling).
func (am *AuthManager) SesssionTimeoutIsRolling() bool {
	return !am.sessionTimeoutIsAbsolute
}

// WithSessionTimeoutProvider sets the session timeout provider.
func (am *AuthManager) WithSessionTimeoutProvider(timeoutProvider func(rc *Ctx) *time.Time) *AuthManager {
	am.SetSessionTimeoutProvider(timeoutProvider)
	return am
}

// SetSessionTimeoutProvider sets the session to expire with a given the given timeout provider.
func (am *AuthManager) SetSessionTimeoutProvider(timeoutProvider func(rc *Ctx) *time.Time) {
	am.sessionTimeoutProvider = timeoutProvider
}

// SessionTimeoutProvider returns the session timeout provider.
func (am *AuthManager) SessionTimeoutProvider() func(rc *Ctx) *time.Time {
	return am.sessionTimeoutProvider
}

// WithCookiesHTTPSOnly sets if we should issue cookies with the HTTPS flag on.
func (am *AuthManager) WithCookiesHTTPSOnly(isHTTPSOnly bool) *AuthManager {
	am.cookieHTTPSOnly = isHTTPSOnly
	return am
}

// SetCookieHTTPSOnly overrides defaults when determining if we should use the HTTPS only cooikie option.
// The default depends on the app configuration (if tls is configured and enabled).
func (am *AuthManager) SetCookieHTTPSOnly(isHTTPSOnly bool) {
	am.cookieHTTPSOnly = isHTTPSOnly
}

// CookiesHTTPSOnly returns if the cookie is for only https connections.
func (am *AuthManager) CookiesHTTPSOnly() bool {
	return am.cookieHTTPSOnly
}

// WithCookieName sets the cookie name.
func (am *AuthManager) WithCookieName(paramName string) *AuthManager {
	am.SetCookieName(paramName)
	return am
}

// SetCookieName sets the session cookie name.
func (am *AuthManager) SetCookieName(paramName string) {
	am.cookieName = paramName
}

// CookieName returns the session param name.
func (am *AuthManager) CookieName() string {
	return am.cookieName
}

// WithCookiePath sets the cookie path.
func (am *AuthManager) WithCookiePath(path string) *AuthManager {
	am.SetCookiePath(path)
	return am
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

// WithSecureCookieName sets the secure cookie name.
func (am *AuthManager) WithSecureCookieName(paramName string) *AuthManager {
	am.SetSecureCookieName(paramName)
	return am
}

// SetSecureCookieName sets the session param name.
func (am *AuthManager) SetSecureCookieName(paramName string) {
	am.secureCookieName = paramName
}

// SecureCookieName returns the session param name.
func (am *AuthManager) SecureCookieName() string {
	return am.secureCookieName
}

// WithPersistHandler sets the persist handler.
func (am *AuthManager) WithPersistHandler(handler func(*Ctx, *Session, State) error) *AuthManager {
	am.SetPersistHandler(handler)
	return am
}

// SetPersistHandler sets the persist handler.
// It must be able to both create sessions and update sessions if the expiry changes.
func (am *AuthManager) SetPersistHandler(handler func(*Ctx, *Session, State) error) {
	am.persistHandler = handler
}

// PersistHandler returns the persist handler.
func (am *AuthManager) PersistHandler() func(*Ctx, *Session, State) error {
	return am.persistHandler
}

// WithFetchHandler sets the fetch handler.
func (am *AuthManager) WithFetchHandler(handler func(sessionID string, state State) (*Session, error)) *AuthManager {
	am.fetchHandler = handler
	return am
}

// SetFetchHandler sets the fetch handler.
func (am *AuthManager) SetFetchHandler(handler func(sessionID string, state State) (*Session, error)) {
	am.fetchHandler = handler
}

// FetchHandler returns the fetch handler.
// It is used in `VerifySession` to satisfy session cache misses.
func (am *AuthManager) FetchHandler() func(sessionID string, state State) (*Session, error) {
	return am.fetchHandler
}

// WithRemoveHandler sets the remove handler.
func (am *AuthManager) WithRemoveHandler(handler func(sessionID string, state State) error) *AuthManager {
	am.SetRemoveHandler(handler)
	return am
}

// SetRemoveHandler sets the remove handler.
// It should remove a session from the backing store by a string sessionID.
func (am *AuthManager) SetRemoveHandler(handler func(sessionID string, state State) error) {
	am.removeHandler = handler
}

// RemoveHandler returns the remove handler.
// It is used in validate session if the session is found to be invalid.
func (am *AuthManager) RemoveHandler() func(sessionID string, state State) error {
	return am.removeHandler
}

// WithValidateHandler sets the validate handler.
func (am *AuthManager) WithValidateHandler(handler func(*Session, State) error) *AuthManager {
	am.SetValidateHandler(handler)
	return am
}

// SetValidateHandler sets the validate handler.
// This is an optional handler that will evaluate the session when verifying requests that are session aware.
func (am *AuthManager) SetValidateHandler(handler func(*Session, State) error) {
	am.validateHandler = handler
}

// ValidateHandler returns the validate handler.
func (am *AuthManager) ValidateHandler() func(*Session, State) error {
	return am.validateHandler
}

// WithLoginRedirectHandler sets the login redirect handler.
func (am *AuthManager) WithLoginRedirectHandler(handler func(*Ctx) *url.URL) *AuthManager {
	am.SetLoginRedirectHandler(handler)
	return am
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template, provided one is configured.
func (am *AuthManager) SetLoginRedirectHandler(handler func(*Ctx) *url.URL) {
	am.loginRedirectHandler = handler
}

// LoginRedirectHandler returns the login redirect handler.
func (am *AuthManager) LoginRedirectHandler() func(*Ctx) *url.URL {
	return am.loginRedirectHandler
}

// SessionCache returns the session cache.
func (am *AuthManager) SessionCache() *SessionCache {
	return am.sessionCache
}

// WithLogger sets the intance logger and returns a reference.
func (am *AuthManager) WithLogger(log *logger.Logger) *AuthManager {
	am.log = log
	return am
}

// SetLogger sets the intance logger.
func (am *AuthManager) SetLogger(log *logger.Logger) {
	am.log = log
}

// Logger returns the instance logger.
func (am *AuthManager) Logger() *logger.Logger {
	return am.log
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// GenerateSessionTimeout returns the absolute time the cookie would expire.
func (am *AuthManager) GenerateSessionTimeout(context *Ctx) *time.Time {
	if am.sessionTimeout > 0 {
		return util.OptionalTime(time.Now().UTC().Add(am.sessionTimeout))
	} else if am.sessionTimeoutProvider != nil {
		return am.sessionTimeoutProvider(context)
	}
	return nil
}

func (am *AuthManager) shouldIssueSecureSesssionID() bool {
	return len(am.secret) > 0
}

func (am AuthManager) shouldUpdateSessionExpiry() bool {
	return am.SesssionTimeoutIsRolling() && (am.sessionTimeout > 0 || am.sessionTimeoutProvider != nil)
}

// CreateSessionID creates a new session id.
func (am AuthManager) createSessionID() string {
	return NewSessionID()
}

// CreateSecureSessionID creates a secure session id.
func (am AuthManager) createSecureSessionID(sessionID string) (string, error) {
	return EncodeSignSessionID(sessionID, am.secret)
}

// InjectCookie injects a session cookie into the context.
func (am *AuthManager) injectCookie(ctx *Ctx, name, value string, expire *time.Time) {
	path := am.CookiePath()
	https := am.CookiesHTTPSOnly()
	ctx.WriteNewCookie(name, value, expire, path, https)
}

// readParam reads a param from a given request context from either the cookies or headers.
func (am *AuthManager) readParam(name string, ctx *Ctx) string {
	if cookie := ctx.GetCookie(name); cookie != nil {
		return cookie.Value
	}
	return ""
}

// ReadSessionID reads a session id from a given request context.
func (am *AuthManager) readSessionID(ctx *Ctx) string {
	return am.readParam(am.CookieName(), ctx)
}

// ReadSecureSessionID reads a secure session id from a given request context.
func (am *AuthManager) readSecureSessionID(ctx *Ctx) string {
	return am.readParam(am.SecureCookieName(), ctx)
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

	secureSessionIDDecoded, err := Base64Decode(secureSessionID)
	if err != nil {
		return ErrSecureSessionIDInvalid
	}

	signedSessionID, err := SignSessionID(sessionID, am.secret)
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
	if am.log != nil && err != nil {
		am.log.Error(err)
	}
	return err
}
