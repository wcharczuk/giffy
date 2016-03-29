package auth

import (
	"github.com/blendlabs/spiffy"
	"github.com/julienschmidt/httprouter"
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
type SessionAwareControllerAction func(session *Session, ctx *web.HTTPContext) web.ControllerResult

// APISessionAwareAction inserts the session into the context.
func APISessionAwareAction(action SessionAwareControllerAction) httprouter.Handle {
	return SessionAwareAction(web.ProviderAPI, action)
}

// APISessionRequiredAction is an action that requires the user to be logged in.
func APISessionRequiredAction(action SessionAwareControllerAction) httprouter.Handle {
	return SessionRequiredAction(web.ProviderAPI, action)
}

// ViewSessionAwareAction inserts the session into the context.
func ViewSessionAwareAction(action SessionAwareControllerAction) httprouter.Handle {
	return SessionAwareAction(web.ProviderView, action)
}

// ViewSessionRequiredAction is an action that requires the user to be logged in.
func ViewSessionRequiredAction(action SessionAwareControllerAction) httprouter.Handle {
	return SessionRequiredAction(web.ProviderView, action)
}

// SessionAwareAction injects the current session (if there is one) into the middleware.
// CAVEAT; we lock on session, so there cannot be multiple concurrent session aware requests (!!).
func SessionAwareAction(resultProvider web.HTTPResultProvider, action SessionAwareControllerAction) httprouter.Handle {
	return web.Render(func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := VerifySession(sessionID)
			if err != nil {
				return resultProvider.InternalError(err)
			}
			if session != nil {
				session.Lock()
				defer session.Unlock()
			}

			return action(session, ctx)
		}
		return action(nil, ctx)
	})
}

// SessionRequiredAction is an action that requires session.
// CAVEAT; we lock on session, so there cannot be multiple concurrent session aware requests (!!).
func SessionRequiredAction(resultProvider web.HTTPResultProvider, action SessionAwareControllerAction) httprouter.Handle {
	return web.Render(func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
		if len(sessionID) == 0 {
			return resultProvider.NotAuthorized()
		}

		session, sessionErr := VerifySession(sessionID)
		if sessionErr != nil {
			return resultProvider.InternalError(sessionErr)
		}
		if session == nil {
			return resultProvider.NotAuthorized()
		}
		if session.User.IsBanned {
			return resultProvider.NotAuthorized()
		}

		session.Lock()
		defer session.Unlock()

		return action(session, ctx)
	})
}
