package webutil

import (
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/model"
)

const (
	// SessionStateUserKey is the key we store the user in the session state.
	SessionStateUserKey = "User"
)

// GetUser returns the user state from a session.
func GetUser(session *web.Session) *model.User {
	if session == nil {
		return nil
	}
	if userData, hasUser := session.State[SessionStateUserKey]; hasUser {
		if user, isUser := userData.(*model.User); isUser {
			return user
		}
	}
	return nil
}

// SetUser stores a user on a session.
func SetUser(session *web.Session, user *model.User) {
	if session.State == nil {
		session.State = map[string]interface{}{}
	}
	session.State[SessionStateUserKey] = user
}
