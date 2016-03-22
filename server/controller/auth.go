package controller

import (
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// Auth is the main controller for the app.
type Auth struct{}

func (ac Auth) oauthAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session != nil {
		cu := &viewmodel.CurrentUser{}
		cu.SetFromUser(session.User)
		return ctx.View.View("login_complete", loginCompleteArguments{CurrentUser: util.SerializeJSON(cu)})
	}

	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.View.BadRequest("`code` parameter missing, cannot continue")
	}

	var oa external.GoogleOAuthResponse
	err := core.NewExternalRequest().
		AsPost().
		WithScheme("https").
		WithHost("accounts.google.com").
		WithPath("o/oauth2/token").
		WithPostData("client_id", core.ConfigGoogleClientID()).
		WithPostData("client_secret", core.ConfigGoogleSecret()).
		WithPostData("grant_type", "authorization_code").
		WithPostData("redirect_uri", viewmodel.OAuthRedirectURI(ctx.Request)).
		WithPostData("code", code).FetchJSONToObject(&oa)

	if err != nil {
		return ctx.View.InternalError(err)
	}

	profile, err := external.FetchGoogleProfile(oa.AccessToken)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	existingUser, err := model.GetUserByUsername(profile.Email, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	var userID int64
	var sessionID string

	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		user := profile.User()
		err = spiffy.DefaultDb().Create(user)
		if err != nil {
			return ctx.View.InternalError(err)
		}
		userID = user.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, auth.OAuthProviderGoogle, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	//save the credentials
	newCredentials := model.NewUserAuth(userID, oa.AccessToken, oa.IDToken)
	newCredentials.Provider = auth.OAuthProviderGoogle
	err = spiffy.DefaultDb().Create(newCredentials)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	// set up the session
	userSession := model.NewUserSession(userID)
	err = spiffy.DefaultDb().Create(userSession)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	sessionID = userSession.SessionID

	auth.SessionState().Add(userID, sessionID)
	ctx.SetCookie(auth.SessionParamName, sessionID, nil, "/")

	currentUser, err := model.GetUserByID(userID, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	cu := &viewmodel.CurrentUser{}
	cu.SetFromUser(currentUser)

	return ctx.View.View("login_complete", loginCompleteArguments{CurrentUser: util.SerializeJSON(cu)})
}

type loginCompleteArguments struct {
	CurrentUser string `json:"current_user"`
}

func (ac Auth) logoutAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session == nil {
		return ctx.Redirect("/")
	}

	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.View.InternalError(err)
	}
	ctx.ExpireCookie(auth.SessionParamName)

	return ctx.Redirect("/")
}

// Register registers the controllers routes.
func (ac Auth) Register(router *httprouter.Router) {
	router.GET("/oauth", auth.ViewSessionAwareAction(ac.oauthAction))
	router.GET("/logout", auth.ViewSessionAwareAction(ac.logoutAction))
	router.POST("/logout", auth.ViewSessionAwareAction(ac.logoutAction))
}
