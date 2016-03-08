package auth

import (
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func Register(username, password string) (*model.UserAuth, *model.UserSession, error) {

	return nil, nil, nil
}

// Login authenticates a session.
func Login(username, password string) (*model.UserSession, error) {
	authToken, authTokenErr := core.Encrypt(core.ConfigKey(), username)
	if authTokenErr != nil {
		return nil, authTokenErr
	}

	authSecret, authSecretErr := core.Encrypt(core.ConfigKey(), password)
	if authSecretErr != nil {
		return nil, authSecretErr
	}

	if err != nil {
		return nil, err
	}

	if any {
        session := model.NewUserSession(userID int64)
	}

	return nil, exception.New("Invalid login credentials.")
}

// Logout un-authenticates a session.
func Logout(userID int64, sessionID string) error {
	return spiffy.DefaultDb().Exec("DELETE FROM user_session where user_id = $1 and session_id = $2", userID, sessionID)
}

// VerifySession checks a sessionID to see if it's valid.
func VerifySession(sessionID string) (*model.UserSession, error) {
	var session model.UserSession
	sessionErr := spiffy.DefaultDb().GetByID(&session, sessionID)
	return &session, sessionErr
}
