package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	// OAuthProviderGoogle is the google auth provider.
	OAuthProviderGoogle = "google"
	// OAuthProviderFacebook is the facebook auth provider.
	OAuthProviderFacebook = "facebook"
	// OAuthProviderSlack is the google auth provider.
	OAuthProviderSlack = "slack"
)

const (
	// SessionStateUserKey is the key we store the user in the session state.
	SessionStateUserKey = "User"
)

// Auth is the main controller for the app.
type Auth struct {
	Log    logger.Log
	OAuth  *oauth.Manager
	Config *config.Giffy
	Model  *model.Manager
}

func (ac Auth) middleware(extra ...web.Middleware) []web.Middleware {
	return append(extra, []web.Middleware{
		web.SessionAware,
		web.ViewProviderAsDefault,
	}...)
}

// Register registers the controllers routes.
func (ac Auth) Register(app *web.App) {
	app.Auth.CookieDefaults.Name = "giffy"

	app.Auth.LoginRedirectHandler = ac.loginRedirect
	app.Auth.FetchHandler = ac.fetchSession
	app.Auth.PersistHandler = ac.persistSession
	app.Auth.RemoveHandler = ac.removeSession

	app.GET("/oauth/google", ac.oauthGoogleAction)
	app.GET("/oauth/slack", ac.oauthSlackAction, ac.middleware()...)
	app.GET("/logout", ac.logoutAction, ac.middleware()...)
	app.POST("/logout", ac.logoutAction, ac.middleware()...)
}

// loginRedirect returns the login redirect.
// This is used when a client tries to access a session required route and isn't authed.
// It should generally point to the login page.
func (ac Auth) loginRedirect(ctx *web.Ctx) *url.URL {
	from := ctx.Request.URL
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

// fetchSession fetches a session from the db.
// Returning `nil` for the session represents a logged out state, and will trigger
// an auth redirect (if one is provided) or a 403 (not authorized) result.
func (ac Auth) fetchSession(ctx context.Context, sessionID string) (*web.Session, error) {
	var session model.UserSession
	_, err := ac.Model.Invoke(ctx).Get(&session, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsZero() {
		return nil, nil
	}

	// check if the user exists in the database
	var dbUser model.User
	_, err = ac.Model.Invoke(ctx).Get(&dbUser, session.UserID)
	if err != nil {
		return nil, err
	}

	if dbUser.IsZero() {
		return nil, nil
	}

	newSession := &web.Session{
		CreatedUTC: session.TimestampUTC,
		SessionID:  sessionID,
		UserID:     strconv.FormatInt(session.UserID, 10),
	}
	SetUser(newSession, &dbUser)
	return newSession, nil
}

// persistSession saves a session to the db.
// It is called when the user logs into the session manager, and allows sessions to persist
// across server restarts.
func (ac Auth) persistSession(ctx context.Context, session *web.Session) error {
	dbSession := &model.UserSession{
		SessionID:    session.SessionID,
		TimestampUTC: session.CreatedUTC,
		UserID:       parseInt64(session.UserID),
	}

	return ac.Model.Invoke(ctx).CreateIfNotExists(dbSession)
}

// removeSession removes a session from the db.
// It is called when the user logs out, and removes their session from the db so it isn't
// returned by `FetchSession`
func (ac Auth) removeSession(ctx context.Context, sessionID string) error {
	var session model.UserSession
	_, err := ac.Model.Invoke(context.TODO()).Get(&session, sessionID)
	if err != nil {
		return err
	}
	_, err = ac.Model.Invoke(ctx).Delete(session)
	if err != nil {
		return err
	}
	return nil
}

func (ac Auth) oauthSlackAction(r *web.Ctx) web.Result {
	code := web.StringValue(r.Param("code"))
	if len(code) == 0 {
		return r.Views.BadRequest(fmt.Errorf("`code` parameter missing, cannot continue"))
	}

	res, err := external.SlackOAuth(code, ac.Config)
	if err != nil {
		return r.Views.InternalError(err)
	}

	if !res.OK {
		return r.Views.InternalError(fmt.Errorf("Slack Error: %s", res.Error))
	}

	auth, err := external.FetchSlackProfile(res.AccessToken, ac.Config)
	if err != nil {
		return r.Views.InternalError(err)
	}

	existingTeam, err := ac.Model.GetSlackTeamByTeamID(r.Context(), auth.TeamID)
	if err != nil {
		return r.Views.InternalError(err)
	}

	if existingTeam.IsZero() {
		team := model.NewSlackTeam(auth.TeamID, auth.Team, auth.UserID, auth.User)
		err = ac.Model.Invoke(r.Context()).Create(team)
		if err != nil {
			return r.Views.InternalError(err)
		}
	}
	return web.RedirectWithMethodf(http.MethodGet, "/slack/complete")
}

func (ac Auth) mapGoogleUser(profile *oauth.Profile) *model.User {
	user := model.NewUser(profile.Email)
	user.EmailAddress = profile.Email
	user.IsEmailVerified = profile.VerifiedEmail
	user.FirstName = profile.GivenName
	user.LastName = profile.FamilyName
	return user
}

func (ac Auth) oauthGoogleAction(r *web.Ctx) web.Result {
	res, err := ac.OAuth.Finish(r.Request)
	if err != nil {
		logger.MaybeWarning(ac.Log, err)
		return r.Views.NotAuthorized()
	}
	prototypeUser := ac.mapGoogleUser(&res.Profile)
	return ac.finishOAuthLogin(r, OAuthProviderGoogle, res.Response.AccessToken, "", prototypeUser)
}

func (ac Auth) finishOAuthLogin(r *web.Ctx, provider, authToken, authSecret string, prototypeUser *model.User) web.Result {
	existingUser, err := ac.Model.GetUserByUsername(r.Context(), prototypeUser.Username)
	if err != nil {
		return r.Views.InternalError(err)
	}

	var userID int64
	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		err = ac.Model.Invoke(r.Context()).Create(prototypeUser)
		if err != nil {
			return r.Views.InternalError(err)
		}
		userID = prototypeUser.ID
	} else {
		userID = existingUser.ID
	}

	err = ac.Model.DeleteUserAuthForProvider(r.Context(), userID, provider)
	if err != nil {
		return r.Views.InternalError(err)
	}

	//save the credentials
	newCredentials, err := model.NewUserAuth(userID, authToken, authSecret, ac.Config.GetEncryptionKey())
	if err != nil {
		return r.Views.InternalError(err)
	}

	newCredentials.Provider = provider

	err = ac.Model.Invoke(r.Context()).Create(newCredentials)
	if err != nil {
		return r.Views.InternalError(err)
	}

	session, err := r.Auth.Login(strconv.FormatInt(userID, 10), r)
	if err != nil {
		return r.Views.InternalError(err)
	}

	currentUser, err := ac.Model.GetUserByID(r.Context(), userID)
	if err != nil {
		return r.Views.InternalError(err)
	}

	SetUser(session, currentUser)

	cu := &viewmodel.CurrentUser{
		IsLoggedIn:    true,
		SlackLoginURL: external.SlackAuthURL(ac.Config),
	}

	cu.SetFromUser(currentUser)
	return r.Views.View("login_complete", loginCompleteArguments{CurrentUser: toJSON(cu)})
}

type loginCompleteArguments struct {
	CurrentUser string `json:"current_user"`
}

func (ac Auth) logoutAction(r *web.Ctx) web.Result {
	session := r.Session
	if session == nil {
		return web.Redirectf("/")
	}

	err := r.Auth.Logout(r)
	if err != nil {
		return r.Views.InternalError(err)
	}

	return web.Redirectf("/")
}
