package webutil

import (
	"database/sql"
	"fmt"
	"net/url"

	web "github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"
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
func LoginRedirect(from *url.URL) *url.URL {
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
func FetchSession(sessionID string, tx *sql.Tx) (*web.Session, error) {
	var session model.UserSession
	err := spiffy.DefaultDb().GetByIDInTx(&session, tx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsZero() {
		return nil, nil
	}

	// check if the user exists in the database
	var dbUser model.User
	err = spiffy.DefaultDb().GetByIDInTx(&dbUser, tx, session.UserID)
	if err != nil {
		return nil, err
	}

	if dbUser.IsZero() {
		return nil, nil
	}

	newSession := &web.Session{
		CreatedUTC: session.TimestampUTC,
		SessionID:  sessionID,
		UserID:     int64(session.UserID),
	}
	SetUser(newSession, &dbUser)
	return newSession, nil
}

// PersistSession saves a session to the db.
// It is called when the user logs into the session manager, and allows sessions to persist
// across server restarts.
func PersistSession(context *web.Ctx, session *web.Session, tx *sql.Tx) error {
	dbSession := &model.UserSession{
		SessionID:    session.SessionID,
		TimestampUTC: session.CreatedUTC,
		UserID:       session.UserID,
	}

	return spiffy.DefaultDb().CreateIfNotExistsInTx(dbSession, tx)
}

// RemoveSession removes a session from the db.
// It is called when the user logs out, and removes their session from the db so it isn't
// returned by `FetchSession`
func RemoveSession(sessionID string, tx *sql.Tx) error {
	var session model.UserSession
	err := spiffy.DB().GetByIDInTx(&session, tx, sessionID)
	if err != nil {
		return err
	}
	return spiffy.DB().DeleteInTx(session, tx)
}
