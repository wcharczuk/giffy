package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	// OAuthProviderGoogle is the only auth provider we use right now.
	OAuthProviderGoogle = "google"
)

func getImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(image)
}

func getImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(images)
}

func getImagesForTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}

	results, err := model.GetImagesForTagID(tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func getTagsForImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}

	results, err := model.GetTagsForImageID(image.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func getTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(tag)
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

	image, err := createImageFromFile(session.UserID, files[0])
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(image)
}

func createTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	tag := &model.Tag{}
	err := ctx.PostBodyAsJSON(tag)
	if err != nil {
		return ctx.BadRequest(err.Error())
	}

	if len(tag.TagValue) == 0 {
		return ctx.BadRequest("`tag_value` must be set.")
	}

	tagValue := strings.ToLower(tag.TagValue)

	//check if the tag exists first
	existingTag, err := model.GetTagByValue(tagValue, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if !existingTag.IsZero() {
		return ctx.JSON(existingTag)
	}

	tag.UUID = core.UUIDv4().ToShortString()
	tag.CreatedUTC = time.Now().UTC()
	tag.CreatedBy = session.UserID
	tag.TagValue = strings.ToLower(tag.TagValue)

	err = spiffy.DefaultDb().Create(tag)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(tag)
}

func deleteImage(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	imageUUID := ctx.RouteParameter("image_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}
	if !currentUser.IsModerator && image.CreatedBy != currentUser.ID {
		return ctx.NotAuthorized()
	}

	err = model.DeleteImageByID(image.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func deleteTag(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	tagUUID := ctx.RouteParameter("tag_id")

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return ctx.NotAuthorized()
	}

	err = model.DeleteTagByID(tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func upvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return voteResult(true, session, ctx)
}

func downvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return voteResult(false, session, ctx)
}

func voteResult(isUpvote bool, session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	statusCode, err := vote(session, ctx, isUpvote)
	if statusCode == http.StatusOK {
		return ctx.OK()
	} else if statusCode == http.StatusNotFound {
		return ctx.NotFound()
	} else if statusCode == http.StatusInternalServerError && err != nil {
		return ctx.InternalError(err)
	} else if statusCode == http.StatusBadRequest {
		if err != nil {
			if ex := exception.AsException(err); ex != nil {
				return ctx.BadRequest(ex.Message())
			}
			return ctx.BadRequest(err.Error())
		}
		return ctx.BadRequest("There was an issue voting.")
	}
	return ctx.InternalError(exception.New("There was an issue voting."))
}

func vote(session *auth.Session, ctx *web.HTTPContext, isUpvote bool) (int, error) {
	tx, err := spiffy.DefaultDb().Begin()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")
	userID := session.UserID

	tag, err := model.GetTagByUUID(tagUUID, tx)
	if err != nil {
		spiffy.DefaultDb().Rollback(tx)
		return http.StatusInternalServerError, err
	}

	if tag.IsZero() {
		return http.StatusNotFound, exception.New("`tag_id` not found.")
	}

	image, err := model.GetImageByUUID(imageUUID, tx)
	if err != nil {
		return http.StatusInternalServerError, exception.WrapMany(err, spiffy.DefaultDb().Rollback(tx))
	}

	if image.IsZero() {
		return http.StatusNotFound, exception.New("`image_id` not found.")
	}

	existingUserVote, err := model.GetUserVoteForImageAndTag(userID, image.ID, tag.ID, tx)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if !existingUserVote.IsZero() {
		return http.StatusBadRequest, exception.New("Already voted for this image and tag.")
	}

	err = model.CreateOrIncrementVote(userID, image.ID, tag.ID, isUpvote, tx)
	if err != nil {
		return http.StatusInternalServerError, exception.WrapMany(err, spiffy.DefaultDb().Rollback(tx))
	}

	return http.StatusOK, spiffy.DefaultDb().Commit(tx)
}

func searchAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.QueryImages(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func oauthAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func uploadImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.View("upload_image", nil)
}

func uploadImageCompleteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return ctx.BadRequest("No files posted.")
	}

	if len(files) > 1 {
		return ctx.BadRequest("Too many files posted.")
	}

	image, err := createImageFromFile(session.UserID, files[0])
	if err != nil {
		return ctx.InternalError(err)
	}
	if image == nil {
		return ctx.InternalError(exception.New("Nil image returned from `createImageFromFile`."))
	}
	tagValue := strings.ToLower(ctx.Param("tag_value"))

	existingTag, err := model.GetTagByValue(tagValue, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	var tagID int64
	if existingTag.IsZero() {
		newTag := model.NewTag(session.UserID, tagValue)
		err = spiffy.DefaultDb().Create(newTag)
		if err != nil {
			return ctx.InternalError(err)
		}
		tagID = newTag.ID

		err = model.UpdateImageDisplayName(image.ID, tagValue, nil)
		if err != nil {
			return ctx.InternalError(err)
		}
	} else {
		tagID = existingTag.ID
	}

	// automatically vote for the tag <==> image
	err = model.CreateOrIncrementVote(session.UserID, image.ID, tagID, true, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.View("upload_image_complete", image)
}

func indexAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("server/_static/index.html")
}

func logoutAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.InternalError(err)
	}
	ctx.ExpireCookie(auth.SessionParamName)

	return ctx.Redirect("/")
}

func getCurrentUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut(ctx)
		return ctx.JSON(cu)
	}
	user, userErr := model.GetUserByID(session.UserID, nil)
	if userErr != nil {
		return ctx.InternalError(userErr)
	}

	cu.SetFromUser(user)
	return ctx.JSON(cu)
}

func getSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	value, hasValue := session.State[key]
	if !hasValue {
		return ctx.NotFound()
	}
	return ctx.JSON(value)
}

func setSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	session.State[key] = ctx.PostBodyAsString()
	return ctx.OK()
}

// createImageFromFile creates and uploads an image from a posted file.
func createImageFromFile(userID int64, file web.PostedFile) (*model.Image, error) {
	newImage, err := model.NewImageFromPostedFile(userID, file)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(file.Contents)
	remoteEntry, err := filecache.UploadFile(buf, filecache.FileType{Extension: newImage.Extension, MimeType: http.DetectContentType(file.Contents)})
	if err != nil {
		return nil, err
	}

	newImage.S3Bucket = remoteEntry.Bucket
	newImage.S3Key = remoteEntry.Key
	newImage.S3ReadURL = fmt.Sprintf("https://s3-us-west-2.amazonaws.com/%s/%s", remoteEntry.Bucket, remoteEntry.Key)

	err = spiffy.DefaultDb().Create(newImage)
	if err != nil {
		return nil, err
	}

	return newImage, nil
}

// Start starts the app.
func Start() {
	core.DBInit()

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	)

	router := httprouter.New()

	router.GET("/", web.ActionHandler(auth.SessionAwareAction(indexAction)))

	//api endpoints
	router.GET("/api/image/:image_id", web.ActionHandler(getImageAction))
	router.GET("/api/images", web.ActionHandler(getImagesAction))
	router.POST("/api/images", web.ActionHandler(auth.SessionRequiredAction(createImageAction)))

	router.GET("/api/images/tag/:tag_id", web.ActionHandler(getImagesForTagAction))
	router.GET("/api/tags/image/:image_id", web.ActionHandler(getTagsForImageAction))

	router.GET("/images/upload", web.ActionHandler(auth.SessionRequiredAction(uploadImageAction)))
	router.POST("/images/upload", web.ActionHandler(auth.SessionRequiredAction(uploadImageCompleteAction)))

	router.GET("/api/tag/:tag_id", web.ActionHandler(getTagAction))
	router.GET("/api/tags", web.ActionHandler(getTagsAction))
	router.POST("/api/tags", web.ActionHandler(auth.SessionRequiredAction(createTagAction)))
	router.DELETE("/api/tag/:tag_id", web.ActionHandler(auth.SessionRequiredAction(deleteTag)))

	router.GET("/api/users", web.ActionHandler(getUsersAction))
	router.POST("/api/upvote/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(upvoteAction)))
	router.POST("/api/downvote/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(downvoteAction)))
	router.GET("/api/search", web.ActionHandler(searchAction))

	//session endpoints
	router.GET("/api/current_user", web.ActionHandler(auth.SessionAwareAction(getCurrentUserAction)))
	router.GET("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(getSessionKeyAction)))
	router.POST("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(setSessionKeyAction)))

	//auth endpoints
	router.GET("/oauth", web.ActionHandler(auth.SessionAwareAction(oauthAction)))
	router.GET("/logout", web.ActionHandler(auth.SessionRequiredAction(logoutAction)))
	router.POST("/logout", web.ActionHandler(auth.SessionRequiredAction(logoutAction)))

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