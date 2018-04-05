package webutil

import (
	"fmt"
	"net/url"

	"github.com/blend/go-sdk/util"
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

// LoginRedirect returns the login redirect.
// This is used when a client tries to access a session required route and isn't authed.
// It should generally point to the login page.
func LoginRedirect(ctx *web.Ctx) *url.URL {
	from := ctx.Request().URL
	if from.Path != "/" {
		return &url.URL{
			Path:     "/",
			RawQuery: fmt.Sprintf("redirect=%s", url.QueryEscape(from.Path)),
		}
	}
	return &url.URL{
		Path: "/",
	}
}

// FetchSession fetches a session from the db.
// Returning `nil` for the session represents a logged out state, and will trigger
// an auth redirect (if one is provided) or a 403 (not authorized) result.
func FetchSession(sessionID string, state web.State) (*web.Session, error) {
	tx := web.TxFromState(state)
	var session model.UserSession
	err := model.DB().GetInTx(&session, tx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsZero() {
		return nil, nil
	}

	// check if the user exists in the database
	var dbUser model.User
	err = model.DB().GetInTx(&dbUser, tx, session.UserID)
	if err != nil {
		return nil, err
	}

	if dbUser.IsZero() {
		return nil, nil
	}

	newSession := &web.Session{
		CreatedUTC: session.TimestampUTC,
		SessionID:  sessionID,
		UserID:     util.String.FromInt64(session.UserID),
	}
	SetUser(newSession, &dbUser)
	return newSession, nil
}

// PersistSession saves a session to the db.
// It is called when the user logs into the session manager, and allows sessions to persist
// across server restarts.
func PersistSession(context *web.Ctx, session *web.Session, state web.State) error {
	tx := web.TxFromState(state)
	dbSession := &model.UserSession{
		SessionID:    session.SessionID,
		TimestampUTC: session.CreatedUTC,
		UserID:       util.Parse.Int64(session.UserID),
	}

	return model.DB().CreateIfNotExistsInTx(dbSession, tx)
}

// RemoveSession removes a session from the db.
// It is called when the user logs out, and removes their session from the db so it isn't
// returned by `FetchSession`
func RemoveSession(sessionID string, state web.State) error {
	tx := web.TxFromState(state)
	var session model.UserSession
	err := model.DB().GetInTx(&session, tx, sessionID)
	if err != nil {
		return err
	}
	return model.DB().DeleteInTx(session, tx)
}
