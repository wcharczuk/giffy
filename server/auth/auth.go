package auth

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"

	"github.com/wcharczuk/giffy/server/model"
)

const (
	// SessionParamName is the name of the field that needs to have the sessionID on it.
	SessionParamName = "giffy_auth"

	// StateKeySession is the state key for the user session.
	StateKeySession = "__session__"

	// OAuthProviderGoogle is the google auth provider.
	OAuthProviderGoogle = "google"

	// OAuthProviderFacebook is the facebook auth provider.
	OAuthProviderFacebook = "facebook"

	// OAuthProviderSlack is the google auth provider.
	OAuthProviderSlack = "slack"

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2
)

// InjectSession injects the session object into a request context.
func InjectSession(session *Session, context *web.Ctx) {
	context.SetState(StateKeySession, session)
}

// GetSession extracts the session from the web.Ctx
func GetSession(context *web.Ctx) *Session {
	if sessionStorage := context.State(StateKeySession); sessionStorage != nil {
		if session, isSession := sessionStorage.(*Session); isSession {
			return session
		}
	}
	return nil
}

// UserProvider is an object that returns a user.
type UserProvider interface {
	AsUser() *model.User
}

// Login logs a userID in.
func Login(userID int64, context *web.Ctx, tx *sql.Tx) (string, error) {
	userSession := model.NewUserSession(userID)
	err := spiffy.DefaultDb().CreateInTx(userSession, tx)
	if err != nil {
		return "", err
	}
	sessionID := userSession.SessionID
	SessionState().Add(userID, sessionID, tx)
	if context != nil {
		context.WriteNewCookie(SessionParamName, sessionID, util.OptionalTime(time.Now().UTC().AddDate(0, 1, 0)), "/", true)
	}

	return sessionID, nil
}

// Logout un-authenticates a session.
func Logout(userID int64, sessionID string, r *web.Ctx, tx *sql.Tx) error {
	SessionState().Expire(sessionID)
	if r != nil {
		r.ExpireCookie(SessionParamName, "/")
	}
	return model.DeleteUserSession(userID, sessionID, tx)
}

// VerifySession checks a sessionID to see if it's valid.
func VerifySession(sessionID string, tx *sql.Tx) (*Session, error) {
	if SessionState().IsActive(sessionID) {
		return SessionState().Get(sessionID), nil
	}

	session := model.UserSession{}
	sessionErr := spiffy.DefaultDb().GetByIDInTx(&session, tx, sessionID)

	if sessionErr != nil {
		return nil, sessionErr
	}

	if session.IsZero() {
		return nil, nil
	}

	return SessionState().Add(session.UserID, session.SessionID, tx)
}

// SessionAware is an action that injects the session into the context, it acquires a read lock on session.
func SessionAware(action web.Action) web.Action {
	return sessionAware(action, SessionReadLock)
}

// SessionAwareMutating is an action that injects the session into the context and requires a write lock.
func SessionAwareMutating(action web.Action) web.Action {
	return sessionAware(action, SessionReadWriteLock)
}

// SessionAwareLockFree is an action that injects the session into the context without acquiring a lock.
func SessionAwareLockFree(action web.Action) web.Action {
	return sessionAware(action, SessionLockFree)
}

func sessionAware(action web.Action, sessionLockPolicy int) web.Action {
	return func(context *web.Ctx) web.Result {
		if context.DefaultResultProvider() == nil {
			panic("You must provide a content provider as middleware to use `SessionAware`")
		}

		sessionID := context.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := VerifySession(sessionID, context.Tx())
			if err != nil {
				return context.DefaultResultProvider().InternalError(err)
			}

			if session != nil {
				switch sessionLockPolicy {
				case SessionReadLock:
					{
						session.RLock()
						defer session.RUnlock()
						break
					}
				case SessionReadWriteLock:
					{
						session.Lock()
						defer session.Unlock()
						break
					}
				}
			}

			InjectSession(session, context)
		}
		return action(context)
	}
}

// SessionRequired is an action that requires a (valid) session to be present
// or identified in some form on the request, and acquires a read lock on session.
func SessionRequired(action web.Action) web.Action {
	return sessionRequired(action, SessionReadLock)
}

// SessionRequiredMutating is an action that requires the session to present and also requires a write lock.
func SessionRequiredMutating(action web.Action) web.Action {
	return sessionRequired(action, SessionReadWriteLock)
}

// SessionRequiredLockFree is an action that requires the session to present and does not acquire any lock.
func SessionRequiredLockFree(action web.Action) web.Action {
	return sessionRequired(action, SessionLockFree)
}

func sessionRequired(action web.Action, sessionLockPolicy int) web.Action {
	return func(context *web.Ctx) web.Result {
		if context.DefaultResultProvider() == nil {
			panic("You must provide a content provider as middleware to use `SessionRequired`")
		}

		sessionID := context.Param(SessionParamName)
		if len(sessionID) == 0 {
			return context.DefaultResultProvider().NotAuthorized()
		}

		session, sessionErr := VerifySession(sessionID, context.Tx())
		if sessionErr != nil {
			return context.DefaultResultProvider().InternalError(sessionErr)
		}
		if session == nil {
			return context.DefaultResultProvider().NotAuthorized()
		}
		if session.User != nil && session.User.IsBanned {
			return context.DefaultResultProvider().NotAuthorized()
		}

		switch sessionLockPolicy {
		case SessionReadLock:
			{
				session.RLock()
				defer session.RUnlock()
				break
			}
		case SessionReadWriteLock:
			{
				session.Lock()
				defer session.Unlock()
				break
			}
		}

		InjectSession(session, context)
		return action(context)
	}
}
