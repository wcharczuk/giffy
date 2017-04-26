package controller

import (
	"fmt"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"

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
type Auth struct{}

// Register registers the controllers routes.
func (ac Auth) Register(app *web.App) {

	app.Auth().SetLoginRedirectHandler(webutil.LoginRedirect)
	app.Auth().SetFetchHandler(webutil.FetchSession)
	app.Auth().SetPersistHandler(webutil.PersistSession)
	app.Auth().SetRemoveHandler(webutil.RemoveSession)

	app.GET("/oauth/google", ac.oauthGoogleAction, web.SessionAwareMutating, web.ViewProviderAsDefault)
	app.GET("/oauth/facebook", ac.oauthFacebookAction, web.SessionAwareMutating, web.ViewProviderAsDefault)
	app.GET("/oauth/slack", ac.oauthSlackAction, web.SessionAwareMutating, web.ViewProviderAsDefault)
	app.GET("/logout", ac.logoutAction, web.SessionRequiredMutating, web.ViewProviderAsDefault)
	app.POST("/logout", ac.logoutAction, web.SessionRequiredMutating, web.ViewProviderAsDefault)
}

func (ac Auth) oauthSlackAction(r *web.Ctx) web.Result {
	code := r.Param("code")
	if len(code) == 0 {
		return r.View().BadRequest("`code` parameter missing, cannot continue")
	}

	res, err := external.SlackOAuth(code)
	if err != nil {
		return r.View().InternalError(err)
	}

	if !res.OK {
		return r.View().InternalError(fmt.Errorf("Slack Error: %s", res.Error))
	}

	auth, err := external.FetchSlackProfile(res.AccessToken)
	if err != nil {
		return r.View().InternalError(err)
	}

	existingTeam, err := model.GetSlackTeamByTeamID(auth.TeamID)
	if err != nil {
		return r.View().InternalError(err)
	}

	if existingTeam.IsZero() {
		team := model.NewSlackTeam(auth.TeamID, auth.Team, auth.UserID, auth.User)
		err = model.DB().CreateInTx(team, r.Tx())
		if err != nil {
			return r.View().InternalError(err)
		}
	}

	return r.Redirect("/slack/complete")
}

func (ac Auth) oauthGoogleAction(r *web.Ctx) web.Result {
	code := r.Param("code")
	if len(code) == 0 {
		return r.View().BadRequest("`code` parameter missing, cannot continue")
	}

	oa, err := external.GoogleOAuth(code)
	if err != nil {
		return r.View().InternalError(err)
	}

	profile, err := external.FetchGoogleProfile(oa.AccessToken)
	if err != nil {
		return r.View().InternalError(err)
	}

	prototypeUser := profile.AsUser()
	return ac.finishOAuthLogin(r, OAuthProviderGoogle, oa.AccessToken, oa.IDToken, prototypeUser)
}

func (ac Auth) oauthFacebookAction(r *web.Ctx) web.Result {
	code := r.Param("code")
	if len(code) == 0 {
		return r.View().BadRequest("`code` parameter missing, cannot continue")
	}

	oa, err := external.FacebookOAuth(code)
	if err != nil {
		return r.View().InternalError(err)
	}

	profile, err := external.FetchFacebookProfile(oa.AccessToken)
	if err != nil {
		return r.View().InternalError(err)
	}

	if len(profile.Email) == 0 {
		return r.View().BadRequest("Facebook privacy settings restrict email; cannot continue.")
	}

	prototypeUser := profile.AsUser()
	return ac.finishOAuthLogin(r, OAuthProviderGoogle, oa.AccessToken, util.StringEmpty, prototypeUser)
}

func (ac Auth) finishOAuthLogin(r *web.Ctx, provider, authToken, authSecret string, prototypeUser *model.User) web.Result {
	existingUser, err := model.GetUserByUsername(prototypeUser.Username, r.Tx())
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
		external.StatHatUserSignup()
		userID = prototypeUser.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, provider, r.Tx())
	if err != nil {
		return r.View().InternalError(err)
	}

	//save the credentials
	newCredentials, err := model.NewUserAuth(userID, authToken, authSecret)
	if err != nil {
		return r.View().InternalError(err)
	}

	newCredentials.Provider = provider

	err = model.DB().Create(newCredentials)

	if err != nil {
		return r.View().InternalError(err)
	}

	session, err := r.Auth().Login(userID, r)
	if err != nil {
		return r.View().InternalError(err)
	}

	currentUser, err := model.GetUserByID(userID, r.Tx())
	if err != nil {
		return r.View().InternalError(err)
	}

	webutil.SetUser(session, currentUser)

	cu := &viewmodel.CurrentUser{}
	cu.SetFromUser(currentUser)

	return r.View().View("login_complete", loginCompleteArguments{CurrentUser: util.JSON.Serialize(cu)})
}

type loginCompleteArguments struct {
	CurrentUser string `json:"current_user"`
}

func (ac Auth) logoutAction(r *web.Ctx) web.Result {
	session := r.Session()
	if session == nil {
		return r.Redirect("/")
	}

	err := r.Auth().Logout(session, r)
	if err != nil {
		return r.View().InternalError(err)
	}

	return r.Redirect("/")
}
