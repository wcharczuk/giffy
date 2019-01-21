package controller

import (
	"fmt"
	"net/url"

	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
	"github.com/wcharczuk/giffy/server/webutil"
)

const (
	// OAuthProviderGoogle is the google auth provider.
	OAuthProviderGoogle = "google"

	// OAuthProviderFacebook is the facebook auth provider.
	OAuthProviderFacebook = "facebook"

	// OAuthProviderSlack is the google auth provider.
	OAuthProviderSlack = "slack"
)

// Auth is the main controller for the app.
type Auth struct {
	OAuth  *oauth.Manager
	Config *config.Giffy
	Model *model.Manager
}

// Register registers the controllers routes.
func (ac Auth) Register(app *web.App) {
	app.Auth().WithLoginRedirectHandler(ac.loginRedirect)
	app.Auth().WithFetchHandler(ac.fetchSession)
	app.Auth().WithPersistHandler(ac.persistSession)
	app.Auth().WithRemoveHandler(ac.removeSession)

	app.GET("/oauth/google", ac.oauthGoogleAction, web.SessionAwareMutating, web.ViewProviderAsDefault)
	app.GET("/oauth/slack", ac.oauthSlackAction, web.SessionAwareMutating, web.ViewProviderAsDefault)
	app.GET("/logout", ac.logoutAction, web.SessionRequiredMutating, web.ViewProviderAsDefault)
	app.POST("/logout", ac.logoutAction, web.SessionRequiredMutating, web.ViewProviderAsDefault)
}

// loginRedirect returns the login redirect.
// This is used when a client tries to access a session required route and isn't authed.
// It should generally point to the login page.
func (ac Auth) loginRedirect(ctx *web.Ctx) *url.URL {
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

// fetchSession fetches a session from the db.
// Returning `nil` for the session represents a logged out state, and will trigger
// an auth redirect (if one is provided) or a 403 (not authorized) result.
func (ac Auth) fetchSession(sessionID string, state web.State) (*web.Session, error) {
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

// persistSession saves a session to the db.
// It is called when the user logs into the session manager, and allows sessions to persist
// across server restarts.
func (ac Auth) persistSession(context *web.Ctx, session *web.Session, state web.State) error {
	dbSession := &model.UserSession{
		SessionID:    session.SessionID,
		TimestampUTC: session.CreatedUTC,
		UserID:       util.Parse.Int64(session.UserID),
	}

	return model.DB().CreateIfNotExistsInTx(dbSession, tx)
}

// removeSession removes a session from the db.
// It is called when the user logs out, and removes their session from the db so it isn't
// returned by `FetchSession`
func (ac Auth) removeSession(sessionID string, state web.State) error {
	var session model.UserSession
	err := ac.Model.Invoke(context.TODO()).Get(&session, sessionID)
	if err != nil {
		return err
	}
	return model.DB().DeleteInTx(session, tx)
}


func (ac Auth) oauthSlackAction(r *web.Ctx) web.Result {
	code := r.ParamString("code")
	if len(code) == 0 {
		return r.View().BadRequest(fmt.Errorf("`code` parameter missing, cannot continue"))
	}

	res, err := external.SlackOAuth(code, ac.Config)
	if err != nil {
		return r.View().InternalError(err)
	}

	if !res.OK {
		return r.View().InternalError(fmt.Errorf("Slack Error: %s", res.Error))
	}

	auth, err := external.FetchSlackProfile(res.AccessToken, ac.Config)
	if err != nil {
		return r.View().InternalError(err)
	}

	existingTeam, err := model.GetSlackTeamByTeamID(auth.TeamID)
	if err != nil {
		return r.View().InternalError(err)
	}

	if existingTeam.IsZero() {
		team := model.NewSlackTeam(auth.TeamID, auth.Team, auth.UserID, auth.User)
		err = model.DB().CreateInTx(team, web.Tx(r))
		if err != nil {
			return r.View().InternalError(err)
		}
	}

	return r.Redirectf("/slack/complete")
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
	res, err := ac.OAuth.Finish(r.Request())
	if err != nil {
		return r.View().NotAuthorized()
	}
	prototypeUser := ac.mapGoogleUser(res.Profile)
	return ac.finishOAuthLogin(r, OAuthProviderGoogle, res.Response.AccessToken, "", prototypeUser)
}

func (ac Auth) finishOAuthLogin(r *web.Ctx, provider, authToken, authSecret string, prototypeUser *model.User) web.Result {
	existingUser, err := model.GetUserByUsername(prototypeUser.Username, web.Tx(r))
	if err != nil {
		return r.View().InternalError(err)
	}

	var userID int64
	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		err = model.DB().Create(prototypeUser)
		if err != nil {
			return r.View().InternalError(err)
		}
		userID = prototypeUser.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, provider, web.Tx(r))
	if err != nil {
		return r.View().InternalError(err)
	}

	//save the credentials
	newCredentials, err := model.NewUserAuth(userID, authToken, authSecret, ac.Config.GetEncryptionKey())
	if err != nil {
		return r.View().InternalError(err)
	}

	newCredentials.Provider = provider

	err = model.DB().Create(newCredentials)

	if err != nil {
		return r.View().InternalError(err)
	}

	session, err := r.Auth().Login(util.String.FromInt64(userID), r)
	if err != nil {
		return r.View().InternalError(err)
	}

	currentUser, err := model.GetUserByID(userID, web.Tx(r))
	if err != nil {
		return r.View().InternalError(err)
	}

	webutil.SetUser(session, currentUser)

	cu := &viewmodel.CurrentUser{
		IsLoggedIn:    true,
		SlackLoginURL: external.SlackAuthURL(ac.Config),
	}

	cu.SetFromUser(currentUser)
	return r.View().View("login_complete", loginCompleteArguments{CurrentUser: util.JSON.Serialize(cu)})
}

type loginCompleteArguments struct {
	CurrentUser string `json:"current_user"`
}

func (ac Auth) logoutAction(r *web.Ctx) web.Result {
	session := r.Session()
	if session == nil {
		return r.Redirectf("/")
	}

	err := r.Auth().Logout(r)
	if err != nil {
		return r.View().InternalError(err)
	}

	return r.Redirectf("/")
}
