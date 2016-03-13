package auth

import (
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

const (
	// SessionParamName is the name of the field that needs to have the sessionID on it.
	SessionParamName = "giffy_auth"

	// StateKeySession is the state key for the user session.
	StateKeySession = "__session__"
)

// Login authenticates a session.
func Login(token, secret string) (*Session, error) {
	userAuth, userAuthErr := model.GetUserAuthByTokenAndSecret(token, secret, nil)
	if userAuthErr != nil {
		return nil, userAuthErr
	}

	if userAuth.IsZero() { // not authorized ...
		return nil, nil
	}

	userSession := model.NewUserSession(userAuth.UserID)
	err := spiffy.DefaultDb().Create(userSession)
	if err != nil {
		return nil, err
	}

	session := NewSession(userAuth.UserID, userSession.SessionID)

	return session, nil
}

// Logout un-authenticates a session.
func Logout(userID int64, sessionID string) error {
	SessionState().Expire(sessionID)
	return model.DeleteUserSession(userID, sessionID, nil)
}

// VerifySession checks a sessionID to see if it's valid.
func VerifySession(sessionID string) (*Session, error) {
	if SessionState().IsActive(sessionID) {
		return SessionState().Sessions[sessionID], nil
	}

	var session model.UserSession
	sessionErr := spiffy.DefaultDb().GetByID(&session, sessionID)

	if sessionErr != nil {
		return nil, sessionErr
	}

	if session.IsZero() {
		return nil, nil
	}

	return SessionState().Add(session.UserID, session.SessionID), nil
}

// SessionAwareControllerAction is an controller action that also gets the session passed in.
type SessionAwareControllerAction func(session *Session, ctx *web.HTTPContext) web.ControllerResult

// SessionAwareAction inserts the session into the context.
func SessionAwareAction(action SessionAwareControllerAction) web.ControllerAction {
	return func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := VerifySession(sessionID)
			if err != nil {
				return ctx.InternalError(err)
			}
			ctx.SetState(StateKeySession, session)
			return action(session, ctx)
		}
		return action(nil, ctx)
	}
}

// SessionRequiredAction is an action that requires the user to be logged in.
func SessionRequiredAction(action SessionAwareControllerAction) web.ControllerAction {
	return func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
		if len(sessionID) == 0 {
			return ctx.NotAuthorized()
		}

		session, sessionErr := VerifySession(sessionID)
		if sessionErr != nil {
			return ctx.InternalError(sessionErr)
		}

		if session == nil {
			return ctx.NotAuthorized()
		}

		ctx.SetState(StateKeySession, session)
		return action(session, ctx)
	}
}
