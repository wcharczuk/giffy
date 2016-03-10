package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	// AuthParamName is the name of the field that needs to have the sessionID on it.
	AuthParamName = "giffy_auth"
	// StateSession is the state key for the user session.
	StateSession = "__session__"
)

// AuthRequiredAction is an action that requires the user to be logged in.
func AuthRequiredAction(action web.APIControllerAction) web.APIControllerAction {
	return func(ctx *web.APIContext) *web.APIResponse {
		sessionID := ctx.Param(AuthParamName)
		if len(sessionID) == 0 {
			return ctx.NotAuthorized()
		}

		session, sessionErr := auth.VerifySession(sessionID)
		if sessionErr != nil {
			return ctx.InternalError(sessionErr)
		}

		if session == nil {
			return ctx.NotAuthorized()
		}

		ctx.SetState(StateSession, session)

		return action(ctx)
	}
}

func getSession(ctx *web.APIContext) *auth.Session {
	if session := ctx.State(StateSession); session != nil {
		if typed, isTyped := session.(*auth.Session); isTyped {
			return typed
		}
	}
	return nil
}

func getImagesAction(ctx *web.APIContext) *web.APIResponse {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(images)
}

func getTagsAction(ctx *web.APIContext) *web.APIResponse {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(tags)
}

func getUsersAction(ctx *web.APIContext) *web.APIResponse {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(users)
}

func createImageAction(ctx *web.APIContext) *web.APIResponse {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.BadRequest("Problem reading posted file.")
	}

	if len(files) == 0 {
		return ctx.BadRequest("No files posted.")
	}

	//upload file to s3, save it etc.
	return ctx.OK()
}

func createTagAction(ctx *web.APIContext) *web.APIResponse {
	var tag model.Tag
	err := ctx.PostBodyAsJSON(&tag)
	if err != nil {
		return ctx.BadRequest(err.Error())
	}

	err = spiffy.DefaultDb().Create(&tag)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(tag)
}

func createUserAction(ctx *web.APIContext) *web.APIResponse {
	var user model.User
	err := ctx.PostBodyAsJSON(&user)
	if err != nil {
		return ctx.BadRequest(err.Error())
	}
	err = spiffy.DefaultDb().Create(&user)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(user)
}

func upvoteAction(ctx *web.APIContext) *web.APIResponse {
	err := vote(ctx, true)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func downvoteAction(ctx *web.APIContext) *web.APIResponse {
	err := vote(ctx, false)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func vote(ctx *web.APIContext, isUpvote bool) error {
	tx, err := spiffy.DefaultDb().Begin()
	if err != nil {
		return err
	}

	session := getSession(ctx)
	if session == nil {
		return exception.New("User is not logged in.")
	}

	imageID := ctx.RouteParameterInt64("image_id")
	tagID := ctx.RouteParameterInt64("tag_id")
	userID := session.UserID

	err = model.Vote(userID, imageID, tagID, isUpvote, tx)
	if err != nil {
		rollbackErr := spiffy.DefaultDb().Rollback(tx)
		return exception.WrapMany(err, rollbackErr)
	}

	return spiffy.DefaultDb().Commit(tx)
}

func searchAction(ctx *web.APIContext) *web.APIResponse {
	query := ctx.Param("query")
	results, err := model.QueryImages(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(results)
}

type authedResponse struct {
	SessionID string `json:"giffy_auth"`
}

func oauthRedirectURI(r *http.Request) string {
	return fmt.Sprintf("http://%s/oauth", core.ConfigHostname())
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

func oauthAction(ctx *web.APIContext) *web.APIResponse {
	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.BadRequest("`code` parameter missing, cannot continue")
	}

	var oa oauthResponse
	err := request.NewHTTPRequest().AsPost().WithScheme("https").WithHost("accounts.google.com").WithPath("o/oauth2/token").
		WithPostData("client_id", core.ConfigGoogleClientID()).
		WithPostData("client_secret", core.ConfigGoogleSecret()).
		WithPostData("grant_type", "authorization_code").
		WithPostData("redirect_uri", oauthRedirectURI(ctx.Request)).
		WithPostData("code", code).FetchJSONToObject(&oa)

	if err != nil {
		return ctx.InternalError(err)
	}

	creds, err := model.GetUserAuthByTokenAndSecret(oa.AccessToken, oa.IDToken, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if creds.IsZero() {
		// create the user
		user := model.NewUser(core.UUIDv4().ToShortString())
		err = spiffy.DefaultDb().Create(user)
		if err != nil {
			return ctx.InternalError(err)
		}
	}

	return ctx.Redirect("/")
}

func logoutAction(ctx *web.APIContext) *web.APIResponse {
	session := getSession(ctx)
	if session != nil {
		return ctx.NotAuthorized()
	}

	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

var templates = template.Must(template.ParseFiles("server/_views/header.html", "server/_views/footer.html", "server/_views/index.html", "server/_views/login.html"))

func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := templates.ExecuteTemplate(w, "index", viewmodel.Index{Title: "Home"})
	if err != nil {
		fmt.Printf("index: %#v\n", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := templates.ExecuteTemplate(w, "login", viewmodel.Login{Title: "Login", ClientID: core.ConfigGoogleClientID(), Secret: core.ConfigGoogleSecret(), OAUTHRedirectURI: oauthRedirectURI(r)})
	if err != nil {
		fmt.Printf("login: %#v\n", err)
	}
}

func initRouter(router *httprouter.Router) {
	router.GET("/images", web.APIActionHandler(getImagesAction))
	router.GET("/tags", web.APIActionHandler(getTagsAction))
	router.GET("/users", web.APIActionHandler(getUsersAction))

	router.GET("/search", web.APIActionHandler(searchAction))

	router.POST("/images", web.APIActionHandler(createImageAction))
	router.POST("/tags", web.APIActionHandler(createTagAction))
	router.POST("/users", web.APIActionHandler(createUserAction))

	router.POST("/upvote/:image_id/:tag_id", web.APIActionHandler(upvoteAction))
	router.POST("/downvote/:image_id/:tag_id", web.APIActionHandler(downvoteAction))

	router.GET("/oauth", web.APIActionHandler(oauthAction))
	router.POST("/logout", web.APIActionHandler(logoutAction))

	router.GET("/", indexHandler)
	router.GET("/login", loginHandler)

	router.ServeFiles("/static/*filepath", http.Dir("_static"))
}

func main() {
	core.DBInit()

	router := httprouter.New()
	initRouter(router)
	router.NotFound = web.APINotFoundHandler
	router.PanicHandler = web.APIPanicHandler

	bindAddr := fmt.Sprintf(":%s", core.ConfigPort())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	web.Logf("Giffy Server Started, listening on %s", bindAddr)
	log.Fatal(server.ListenAndServe())
}
