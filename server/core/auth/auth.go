package auth

import (
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

// Login authenticates a session.
func Login(token, secret string) (*Session, error) {
	auth, authErr := model.GetUserAuthByTokenAndSecret(token, secret, nil)
	if authErr != nil {
		return nil, authErr
	}

	if auth.IsZero() { // not authorized ...
		return nil, nil
	}

	userSession := model.NewUserSession(auth.UserID)
	err := spiffy.DefaultDb().Create(userSession)
	if err != nil {
		return nil, err
	}

	session := NewSession(auth.UserID, userSession.SessionID)

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
