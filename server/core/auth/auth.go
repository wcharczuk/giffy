package auth

import (
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-web"

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
)

// UserProvider is an object that returns a user.
type UserProvider interface {
	AsUser() *model.User
}

// Logout un-authenticates a session.
func Logout(userID int64, sessionID string) error {
	SessionState().Expire(sessionID)
	return model.DeleteUserSession(userID, sessionID, nil)
}

// VerifySession checks a sessionID to see if it's valid.
func VerifySession(sessionID string) (*Session, error) {
	if SessionState().IsActive(sessionID) {
		return SessionState().Get(sessionID), nil
	}

	session := model.UserSession{}
	sessionErr := spiffy.DefaultDb().GetByID(&session, sessionID)

	if sessionErr != nil {
		return nil, sessionErr
	}

	if session.IsZero() {
		return nil, nil
	}

	return SessionState().Add(session.UserID, session.SessionID)
}

// SessionAwareControllerAction is an controller action that also gets the session passed in.
type SessionAwareControllerAction func(session *Session, r *web.RequestContext) web.ControllerResult

// SessionAwareAction injects the current session (if there is one) into the middleware.
// CAVEAT; we lock on session, so there cannot be multiple concurrent session aware requests (!!).
func SessionAwareAction(providerID int, action SessionAwareControllerAction) web.ControllerAction {
	return func(r *web.RequestContext) web.ControllerResult {
		sessionID := r.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := VerifySession(sessionID)
			if err != nil {
				return r.ResultProvider(providerID).InternalError(err)
			}
			if session != nil {
				session.Lock()
				defer session.Unlock()
			}

			return action(session, r)
		}
		return action(nil, r)
	}
}

// SessionRequiredAction is an action that requires session.
// CAVEAT; we lock on session, so there cannot be multiple concurrent session aware requests (!!).
func SessionRequiredAction(providerID int, action SessionAwareControllerAction) web.ControllerAction {
	return func(r *web.RequestContext) web.ControllerResult {
		sessionID := r.Param(SessionParamName)
		if len(sessionID) == 0 {
			return r.ResultProvider(providerID).NotAuthorized()
		}

		session, sessionErr := VerifySession(sessionID)
		if sessionErr != nil {
			return r.ResultProvider(providerID).InternalError(sessionErr)
		}
		if session == nil {
			return r.ResultProvider(providerID).NotAuthorized()
		}
		if session.User != nil && session.User.IsBanned {
			return r.ResultProvider(providerID).NotAuthorized()
		}

		session.Lock()
		defer session.Unlock()

		return action(session, r)
	}
}
