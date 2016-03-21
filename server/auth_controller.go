package server

import (
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// AuthController is the main controller for the app.
type AuthController struct{}

func (ac AuthController) oauthAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session != nil {
		return ctx.Redirect("/")
	}

	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.BadRequest("`code` parameter missing, cannot continue")
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
		return ctx.InternalError(err)
	}

	profile, err := external.FetchGoogleProfile(oa.AccessToken)
	if err != nil {
		return ctx.InternalError(err)
	}

	existingUser, err := model.GetUserByUsername(profile.Email, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	var userID int64
	var sessionID string

	//create the user if it doesn't exist ...
	if existingUser.IsZero() {
		user := profile.User()
		err = spiffy.DefaultDb().Create(user)
		if err != nil {
			return ctx.InternalError(err)
		}
		userID = user.ID
	} else {
		userID = existingUser.ID
	}

	err = model.DeleteUserAuthForProvider(userID, OAuthProviderGoogle, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	//save the credentials
	newCredentials := model.NewUserAuth(userID, oa.AccessToken, oa.IDToken)
	newCredentials.Provider = OAuthProviderGoogle
	err = spiffy.DefaultDb().Create(newCredentials)
	if err != nil {
		return ctx.InternalError(err)
	}

	// set up the session
	userSession := model.NewUserSession(userID)
	err = spiffy.DefaultDb().Create(userSession)
	if err != nil {
		return ctx.InternalError(err)
	}

	sessionID = userSession.SessionID

	auth.SessionState().Add(userID, sessionID)
	ctx.SetCookie(auth.SessionParamName, sessionID, nil, "/")

	return ctx.Redirect("/")
}

func (ac AuthController) logoutAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.InternalError(err)
	}
	ctx.ExpireCookie(auth.SessionParamName)

	return ctx.Redirect("/")
}

// Register registers the controllers routes.
func (ac AuthController) Register(router *httprouter.Router) {
	router.GET("/oauth", web.ActionHandler(auth.SessionAwareAction(ac.oauthAction)))
	router.GET("/logout", web.ActionHandler(auth.SessionRequiredAction(ac.logoutAction)))
	router.POST("/logout", web.ActionHandler(auth.SessionRequiredAction(ac.logoutAction)))
}
