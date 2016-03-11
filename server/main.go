package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	// SessionParamName is the name of the field that needs to have the sessionID on it.
	SessionParamName = "giffy_auth"

	// StateKeySession is the state key for the user session.
	StateKeySession = "__session__"
)

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

type googleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Gender        string `json:"male"`
	Locale        string `json:"locale"`
	PictureURL    string `json:"picture"`
}

func (gp googleProfile) User() *model.User {
	user := model.NewUser(gp.Email)
	user.EmailAddress = gp.Email
	user.IsEmailVerified = gp.VerifiedEmail
	user.FirstName = gp.GivenName
	user.LastName = gp.FamilyName
	return user
}

// SessionAwareControllerAction is an controller action that also gets the session passed in.
type SessionAwareControllerAction func(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult

// SessionAwareAction inserts the session into the context.
func SessionAwareAction(action SessionAwareControllerAction) web.ControllerAction {
	return func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
		if len(sessionID) != 0 {
			session, err := auth.VerifySession(sessionID)
			if err != nil {
				return ctx.InternalError(err)
			}
			ctx.SetState(StateKeySession, session)
			return action(session, ctx)
		}
		return action(nil, ctx)
	}
}

// AuthRequiredAction is an action that requires the user to be logged in.
func AuthRequiredAction(action SessionAwareControllerAction) web.ControllerAction {
	return func(ctx *web.HTTPContext) web.ControllerResult {
		sessionID := ctx.Param(SessionParamName)
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

		ctx.SetState(StateKeySession, session)
		return action(session, ctx)
	}
}

func activeSession(ctx *web.HTTPContext) *auth.Session {
	if session := ctx.State(StateKeySession); session != nil {
		if typed, isTyped := session.(*auth.Session); isTyped {
			return typed
		}
	}
	return nil
}

func getImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(images)
}

func getTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(tags)
}

func getUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(users)
}

func createImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return ctx.BadRequest("No files posted.")
	}

	images := []*model.Image{}

	//upload file to s3, save it etc.
	for _, f := range files {
		buf := bytes.NewBuffer(f.Contents)

		md5sum := model.ConvertMD5(md5.Sum(f.Contents))
		existing, err := model.ImageMD5Check(md5sum, nil)
		if err != nil {
			return ctx.InternalError(err)
		}

		if !existing.IsZero() {
			images = append(images, existing)
		} else {
			newImage := model.NewImage()
			newImage.MD5 = md5sum
			newImage.CreatedBy = session.UserID
			newImage.UpdatedUTC = time.Now().UTC()
			newImage.UpdatedBy = session.UserID

			imageBuf := bytes.NewBuffer(f.Contents)

			imageMeta, _, err := image.DecodeConfig(imageBuf)
			if err != nil {
				return ctx.InternalError(exception.Wrap(err))
			}

			newImage.DisplayName = f.Key
			newImage.Extension = filepath.Ext(f.Filename)
			newImage.Height = imageMeta.Height
			newImage.Width = imageMeta.Width

			remoteEntry, err := filecache.UploadFile(buf, filecache.FileType{Extension: newImage.Extension, MimeType: http.DetectContentType(f.Contents)})
			if err != nil {
				return ctx.InternalError(err)
			}

			newImage.S3Bucket = remoteEntry.Bucket
			newImage.S3Key = remoteEntry.Key
			newImage.S3ReadURL = fmt.Sprintf("https://s3-us-west-2.amazonaws.com/%s/%s", remoteEntry.Bucket, remoteEntry.Key)

			err = spiffy.DefaultDb().Create(newImage)
			if err != nil {
				return ctx.InternalError(err)
			}

			images = append(images, newImage)
		}
	}

	return ctx.JSON(images)
}

func createTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	var tag model.Tag
	err := ctx.PostBodyAsJSON(&tag)
	if err != nil {
		return ctx.BadRequest(err.Error())
	}

	tag.CreatedBy = session.UserID
	tag.CreatedUTC = time.Now().UTC()

	err = spiffy.DefaultDb().Create(&tag)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(tag)
}

func upvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	err := vote(session, ctx, true)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func downvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	err := vote(session, ctx, false)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func vote(session *auth.Session, ctx *web.HTTPContext, isUpvote bool) error {
	tx, err := spiffy.DefaultDb().Begin()
	if err != nil {
		return err
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

func searchAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.QueryImages(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func getGoogleProfile(accessToken string) (*googleProfile, error) {
	var profile googleProfile
	err := core.NewExternalRequest().AsGet().
		WithURL("https://www.googleapis.com/oauth2/v1/userinfo").
		WithQueryString("alt", "json").
		WithQueryString("access_token", accessToken).
		FetchJSONToObject(&profile)
	return &profile, err
}

func oauthAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session != nil {
		return ctx.Redirect("/")
	}

	code := ctx.Param("code")
	if len(code) == 0 {
		return ctx.BadRequest("`code` parameter missing, cannot continue")
	}

	var oa oauthResponse
	err := core.NewExternalRequest().AsPost().WithScheme("https").WithHost("accounts.google.com").WithPath("o/oauth2/token").
		WithPostData("client_id", core.ConfigGoogleClientID()).
		WithPostData("client_secret", core.ConfigGoogleSecret()).
		WithPostData("grant_type", "authorization_code").
		WithPostData("redirect_uri", oauthRedirectURI(ctx.Request)).
		WithPostData("code", code).FetchJSONToObject(&oa)

	if err != nil {
		return ctx.InternalError(err)
	}

	profile, err := getGoogleProfile(oa.AccessToken)
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

	err = model.DeleteUserAuthForProvider(userID, "google", nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	//save the credentials
	newCredentials := model.NewUserAuth(userID, oa.AccessToken, oa.IDToken)
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
	ctx.SetCookie(SessionParamName, sessionID, nil, "/")
	return ctx.Redirect("/")
}

func indexAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("server/_static/index.html")
}

func loginAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session != nil {
		return ctx.Redirect("/")
	}

	return ctx.View("login", viewmodel.Login{
		Title:            "Login",
		ClientID:         core.ConfigGoogleClientID(),
		Secret:           core.ConfigGoogleSecret(),
		OAUTHRedirectURI: oauthRedirectURI(ctx.Request),
	})
}

func logoutAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.InternalError(err)
	}
	ctx.ExpireCookie(SessionParamName)

	return ctx.Redirect("/")
}

func main() {
	core.DBInit()

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/login.html",
	)

	router := httprouter.New()

	router.GET("/", web.ActionHandler(SessionAwareAction(indexAction)))

	//api endpoints
	router.GET("/api/images", web.ActionHandler(getImagesAction))
	router.POST("/api/images", web.ActionHandler(AuthRequiredAction(createImageAction)))
	router.GET("/api/tags", web.ActionHandler(getTagsAction))
	router.POST("/api/tags", web.ActionHandler(AuthRequiredAction(createTagAction)))
	router.GET("/api/users", web.ActionHandler(getUsersAction))
	router.POST("/api/upvote/:image_id/:tag_id", web.ActionHandler(AuthRequiredAction(upvoteAction)))
	router.POST("/api/downvote/:image_id/:tag_id", web.ActionHandler(AuthRequiredAction(downvoteAction)))
	router.GET("/api/search", web.ActionHandler(searchAction))

	//auth endpoints
	router.GET("/login", web.ActionHandler(SessionAwareAction(loginAction)))
	router.GET("/oauth", web.ActionHandler(SessionAwareAction(oauthAction)))
	router.GET("/logout", web.ActionHandler(AuthRequiredAction(logoutAction)))
	router.POST("/logout", web.ActionHandler(AuthRequiredAction(logoutAction)))

	//static files
	router.ServeFiles("/static/*filepath", http.Dir("server/_static"))

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
