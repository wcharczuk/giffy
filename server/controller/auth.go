package controller

import (
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// Auth is the main controller for the app.
type Auth struct{}

func (ac Auth) oauthSlackAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.View.BadRequest("`code` parameter missing, cannot continue")
	}

	_, err := external.SlackOAuth(code)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	return ctx.Redirect("/")
}

func (ac Auth) oauthGoogleAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.View.BadRequest("`code` parameter missing, cannot continue")
	}

	oa, err := external.GoogleOAuth(code)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	profile, err := external.FetchGoogleProfile(oa.AccessToken)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	prototypeUser := profile.AsUser()
	return ac.finishOAuthLogin(ctx, auth.OAuthProviderGoogle, oa.AccessToken, oa.IDToken, prototypeUser)
}

func (ac Auth) finishOAuthLogin(ctx *web.HTTPContext, provider, authToken, authSecret string, prototypeUser *model.User) web.ControllerResult {
	existingUser, err := model.GetUserByUsername(prototypeUser.Username, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	var userID int64
	var sessionID string

	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		err = spiffy.DefaultDb().Create(prototypeUser)
		if err != nil {
			return ctx.View.InternalError(err)
		}
		userID = prototypeUser.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, provider, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	//save the credentials
	newCredentials := model.NewUserAuth(userID, authToken, authSecret)
	newCredentials.Provider = provider
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
	router.GET("/oauth/google", auth.ViewSessionAwareAction(ac.oauthGoogleAction))
	router.GET("/oauth/slack", auth.ViewSessionAwareAction(ac.oauthSlackAction))
	router.GET("/logout", auth.ViewSessionAwareAction(ac.logoutAction))
	router.POST("/logout", auth.ViewSessionAwareAction(ac.logoutAction))
}
