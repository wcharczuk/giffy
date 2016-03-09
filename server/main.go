package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

const (
	// AuthParamName is the name of the field that needs to have the sessionID on it.
	AuthParamName = "giffy_auth"
	// StateSession is the state key for the user session.
	StateSession = "__session__"
)

// AuthRequiredAction is an action that requires the user to be logged in.
func AuthRequiredAction(action web.APIControllerAction) web.APIControllerAction {
	return func(ctx *web.APIContext) *web.ServiceResponse {
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

func getImagesAction(ctx *web.APIContext) *web.ServiceResponse {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(images)
}

func getTagsAction(ctx *web.APIContext) *web.ServiceResponse {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(tags)
}

func getUsersAction(ctx *web.APIContext) *web.ServiceResponse {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.Content(users)
}

func createImageAction(ctx *web.APIContext) *web.ServiceResponse {
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

func createTagAction(ctx *web.APIContext) *web.ServiceResponse {
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

func createUserAction(ctx *web.APIContext) *web.ServiceResponse {
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

func upvoteAction(ctx *web.APIContext) *web.ServiceResponse {
	err := vote(ctx, true)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func downvoteAction(ctx *web.APIContext) *web.ServiceResponse {
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

func searchAction(ctx *web.APIContext) *web.ServiceResponse {
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

func oauthAction(ctx *web.APIContext) *web.ServiceResponse {
	token := ctx.Param("token")
	secret := ctx.Param("secret")

	session, err := auth.Login(token, secret)
	if err != nil {
		return ctx.InternalError(err)
	}

	if session == nil {
		return ctx.NotAuthorized()
	}

	return ctx.Content(authedResponse{SessionID: session.SessionID})
}

func logoutAction(ctx *web.APIContext) *web.ServiceResponse {
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
