package controller

import (
	"time"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// Auth is the main controller for the app.
type Auth struct{}

func (ac Auth) oauthSlackAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)
	code := r.Param("code")
	if len(code) == 0 {
		return r.View().BadRequest("`code` parameter missing, cannot continue")
	}

	_, err := external.SlackOAuth(code)
	if err != nil {
		return r.View().InternalError(err)
	}

	return r.Redirect("/#/slack/complete")
}

func (ac Auth) oauthGoogleAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

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
	return ac.finishOAuthLogin(r, auth.OAuthProviderGoogle, oa.AccessToken, oa.IDToken, prototypeUser)
}

func (ac Auth) oauthFacebookAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

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
	return ac.finishOAuthLogin(r, auth.OAuthProviderGoogle, oa.AccessToken, util.StringEmpty, prototypeUser)
}

func (ac Auth) finishOAuthLogin(r *web.RequestContext, provider, authToken, authSecret string, prototypeUser *model.User) web.ControllerResult {
	existingUser, err := model.GetUserByUsername(prototypeUser.Username, nil)
	if err != nil {
		return r.View().InternalError(err)
	}

	var userID int64
	var sessionID string

	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		err = spiffy.DefaultDb().Create(prototypeUser)
		if err != nil {
			return r.View().InternalError(err)
		}
		external.StatHatUserSignup()
		userID = prototypeUser.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, provider, nil)
	if err != nil {
		return r.View().InternalError(err)
	}

	//save the credentials
	newCredentials := model.NewUserAuth(userID, authToken, authSecret)
	newCredentials.Provider = provider
	err = spiffy.DefaultDb().Create(newCredentials)
	if err != nil {
		return r.View().InternalError(err)
	}

	// set up the session
	userSession := model.NewUserSession(userID)
	err = spiffy.DefaultDb().Create(userSession)
	if err != nil {
		return r.View().InternalError(err)
	}

	sessionID = userSession.SessionID

	auth.SessionState().Add(userID, sessionID)
	r.SetCookie(auth.SessionParamName, sessionID, util.OptionalTime(time.Now().UTC().AddDate(0, 1, 0)), "/")

	currentUser, err := model.GetUserByID(userID, nil)
	if err != nil {
		return r.View().InternalError(err)
	}

	cu := &viewmodel.CurrentUser{}
	cu.SetFromUser(currentUser)

	return r.View().View("login_complete", loginCompleteArguments{CurrentUser: util.SerializeJSON(cu)})
}

type loginCompleteArguments struct {
	CurrentUser string `json:"current_user"`
}

func (ac Auth) logoutAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if session == nil {
		return r.Redirect("/")
	}

	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return r.View().InternalError(err)
	}
	r.ExpireCookie(auth.SessionParamName)

	return r.Redirect("/")
}

// Register registers the controllers routes.
func (ac Auth) Register(app *web.App) {
	app.GET("/oauth/google", ac.oauthGoogleAction, web.ViewProvider, auth.SessionAware)
	app.GET("/oauth/facebook", ac.oauthFacebookAction, web.ViewProvider, auth.SessionAware)
	app.GET("/oauth/slack", ac.oauthSlackAction, web.ViewProvider, auth.SessionAware)
	app.GET("/logout", ac.logoutAction, web.ViewProvider, auth.SessionAware)
	app.POST("/logout", ac.logoutAction, web.ViewProvider, auth.SessionAware)
}
