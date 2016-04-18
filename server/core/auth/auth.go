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

// InjectSession injects the session object into a request context.
func InjectSession(session *Session, context *web.RequestContext) {
	context.SetState(StateKeySession, session)
}

// GetSession extracts the session from the web.RequestContext
func GetSession(context *web.RequestContext) *Session {
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

// SessionAware is an action that injects the session into the context.
func SessionAware(action web.ControllerAction) web.ControllerAction {
	return func(context *web.RequestContext) web.ControllerResult {
		if context.CurrentProvider() == nil {
			panic("context.CurrentProvider() is nil; make sure to include the correct middleware in the handler registration (i.e. web.APIProvider or web.ViewProvider)")
		}

		sessionID := context.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := VerifySession(sessionID)
			if err != nil {
				return context.CurrentProvider().InternalError(err)
			}
			if session != nil {
				session.Lock()
				defer session.Unlock()
			}

			InjectSession(session, context)
		}
		return action(context)
	}
}

// SessionRequired is an action that requires session.
func SessionRequired(action web.ControllerAction) web.ControllerAction {
	return func(context *web.RequestContext) web.ControllerResult {
		if context.CurrentProvider() == nil {
			panic("context.CurrentProvider() is nil; make sure to include the correct middleware in the handler registration (i.e. web.APIProvider or web.ViewProvider)")
		}

		sessionID := context.Param(SessionParamName)
		if len(sessionID) == 0 {
			return context.CurrentProvider().NotAuthorized()
		}

		session, sessionErr := VerifySession(sessionID)
		if sessionErr != nil {
			return context.CurrentProvider().InternalError(sessionErr)
		}
		if session == nil {
			return context.CurrentProvider().NotAuthorized()
		}
		if session.User != nil && session.User.IsBanned {
			return context.CurrentProvider().NotAuthorized()
		}

		session.Lock()
		defer session.Unlock()

		InjectSession(session, context)
		return action(context)
	}
}
