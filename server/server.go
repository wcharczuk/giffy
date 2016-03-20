package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
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

func searchUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	users, err := model.SearchUsers(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(users)
}

func searchImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchImages(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func searchTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchTags(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func updateUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	if !session.User.IsAdmin {
		return ctx.NotAuthorized()
	}

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if user.IsZero() {
		return ctx.NotFound()
	}

	var postedUser model.User
	err = ctx.PostBodyAsJSON(&postedUser)
	if err != nil {
		return ctx.BadRequest("Post body was not properly formatted.")
	}

	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	postedUser.CreatedUTC = user.CreatedUTC

	if !user.IsAdmin && postedUser.IsAdmin {
		return ctx.BadRequest("Cannot promote user to admin through the UI; this must be done in the db directly.")
	}

	if user.IsAdmin && !postedUser.IsAdmin {
		return ctx.BadRequest("Cannot demote user from admin through the UI; this must be done in the db directly.")
	}

	if postedUser.IsAdmin && postedUser.IsBanned {
		return ctx.BadRequest("Cannot ban admins.")
	}

	if !user.IsModerator && postedUser.IsModerator {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbPromoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsModerator && !postedUser.IsModerator {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbDemoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	}

	if !user.IsBanned && postedUser.IsBanned {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbBan, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsBanned && !postedUser.IsBanned {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbUnban, model.ModerationObjectUser, postedUser.UUID)
	}

	err = spiffy.DefaultDb().Update(&postedUser)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(postedUser)
}

func getRecentModerationLog(ctx *web.HTTPContext) web.ControllerResult {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(moderationLog)
}

func getModerationForUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsModerator {
		return ctx.NotAuthorized()
	}

	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if user.IsZero() {
		return ctx.NotFound()
	}

	actions, err := model.GetModerationsForUser(user.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(actions)
}

func getImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
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

func getRandomImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")

	images, err := model.GetRandomImages(count, nil)
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

	if tag.IsZero() {
		tag, err = model.GetTagByValue(tagUUID, nil)
		if err != nil {
			return ctx.InternalError(err)
		}
		if tag.IsZero() {
			return ctx.NotFound()
		}
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

func getUserAction(ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if user.IsZero() {
		return ctx.NotFound()
	}

	return ctx.JSON(user)
}

func getUserImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if user.IsZero() {
		return ctx.NotFound()
	}

	images, err := model.GetImagesForUserID(user.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(images)
}

func createImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return ctx.BadRequest("No files posted.")
	}

	postedFile := files[0]

	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if !existing.IsZero() {
		return ctx.JSON(existing)
	}

	image, err := createImageFromFile(session.UserID, postedFile)
	if err != nil {
		return ctx.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)

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

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID)
	return ctx.JSON(tag)
}

func deleteImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	//delete from s3 (!!)
	err = filecache.DeleteFile(filecache.NewLocationFromKey(image.S3Key))
	if err != nil {
		return ctx.InternalError(err)
	}

	err = model.DeleteImageByID(image.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func deleteTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID)

	return ctx.OK()
}

func getLinksForImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForImage(image.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(voteSummaries)
}

func getLinksForTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForTag(tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(voteSummaries)
}

func getVotesForUserForImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}
	votes, err := model.GetVotesForUserForImage(session.UserID, image.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(votes)
}

func getVotesForUserForTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}
	votes, err := model.GetVotesForUserForTag(session.UserID, tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(votes)
}

func upvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return voteAction(true, session, ctx)
}

func downvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return voteAction(false, session, ctx)
}

func voteAction(isUpvote bool, session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")
	userID := session.UserID

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}

	existingUserVote, err := model.GetVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return ctx.OK()
	}

	err = model.CreateOrIncrementVote(userID, image.ID, tag.ID, isUpvote, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.OK()
}

func deleteUserVoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")
	userID := session.UserID

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.NotFound()
	}

	tx, err := spiffy.DefaultDb().Begin()

	vote, err := model.GetVote(userID, image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return ctx.InternalError(err)
	}
	if vote.IsZero() {
		tx.Rollback()
		return ctx.NotFound()
	}

	// was it an upvote or downvote
	wasUpvote := vote.IsUpvote

	// adjust the vote summary ...
	voteSummary, err := model.GetVoteSummary(image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return ctx.InternalError(err)
	}

	if wasUpvote {
		voteSummary.VotesFor--
	} else {
		voteSummary.VotesAgainst--
	}

	err = model.SetVoteCount(image.ID, tag.ID, voteSummary.VotesFor, voteSummary.VotesAgainst, tx)
	if err != nil {
		tx.Rollback()
		return ctx.InternalError(err)
	}

	err = model.DeleteVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		tx.Rollback()
		return ctx.InternalError(err)
	}

	tx.Commit()
	return ctx.OK()
}

func deleteLinkAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	if image.IsZero() {
		return ctx.NotFound()
	}

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

	err = model.DeleteVoteSummary(image.ID, tag.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, fmt.Sprintf("image: %s tag: %s", imageUUID, tagUUID))

	return ctx.OK()
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

	postedFile := files[0]

	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if !existing.IsZero() {
		return ctx.View("upload_image_complete", existing)
	}

	image, err := createImageFromFile(session.UserID, postedFile)
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

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
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
	cu.SetFromUser(session.User)
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

func getModerationLogByCountAndOffsetAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")
	offset := ctx.RouteParameterInt("offset")

	log, err := model.GetModerationLogByCountAndOffset(count, offset, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(log)
}

// Init inits the app.
func Init() *httprouter.Router {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	)

	router := httprouter.New()

	router.GET("/", web.ActionHandler(auth.SessionAwareAction(indexAction)))

	//api endpoints
	router.GET("/api/users", web.ActionHandler(getUsersAction))
	router.GET("/api/user/:user_id", web.ActionHandler(getUserAction))
	router.PUT("/api/user/:user_id", web.ActionHandler(auth.SessionRequiredAction(updateUserAction)))

	router.GET("/api/user.images/:user_id", web.ActionHandler(getUserImagesAction))

	router.GET("/api/user.current", web.ActionHandler(auth.SessionAwareAction(getCurrentUserAction)))

	router.GET("/api/images", web.ActionHandler(getImagesAction))
	router.POST("/api/images", web.ActionHandler(auth.SessionRequiredAction(createImageAction)))
	router.GET("/api/images/random/:count", web.ActionHandler(getRandomImagesAction))

	router.GET("/api/image/:image_id", web.ActionHandler(getImageAction))
	router.DELETE("/api/image/:image_id", web.ActionHandler(auth.SessionRequiredAction(deleteImageAction)))

	router.GET("/images/upload", web.ActionHandler(auth.SessionRequiredAction(uploadImageAction)))
	router.POST("/images/upload", web.ActionHandler(auth.SessionRequiredAction(uploadImageCompleteAction)))

	router.GET("/api/tag.images/:tag_id", web.ActionHandler(getImagesForTagAction))
	router.GET("/api/image.tags/:image_id", web.ActionHandler(getTagsForImageAction))

	router.GET("/api/tags", web.ActionHandler(getTagsAction))
	router.POST("/api/tags", web.ActionHandler(auth.SessionRequiredAction(createTagAction)))
	router.GET("/api/tag/:tag_id", web.ActionHandler(getTagAction))
	router.DELETE("/api/tag/:tag_id", web.ActionHandler(auth.SessionRequiredAction(deleteTagAction)))

	router.GET("/api/image.votes/:image_id", web.ActionHandler(getLinksForImageAction))
	router.GET("/api/tag.votes/:tag_id", web.ActionHandler(getLinksForTagAction))

	router.DELETE("/api/link/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(deleteLinkAction)))

	router.GET("/api/user.votes.image/:image_id", web.ActionHandler(auth.SessionRequiredAction(getVotesForUserForImageAction)))
	router.GET("/api/user.votes.tag/:tag_id", web.ActionHandler(auth.SessionRequiredAction(getVotesForUserForTagAction)))
	router.DELETE("/api/user.vote/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(deleteUserVoteAction)))

	router.POST("/api/vote.up/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(upvoteAction)))
	router.POST("/api/vote.down/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(downvoteAction)))

	router.GET("/api/users.search", web.ActionHandler(searchUsersAction))
	router.GET("/api/images.search", web.ActionHandler(searchImagesAction))
	router.GET("/api/tags.search", web.ActionHandler(searchTagsAction))

	router.GET("/api/moderation.log/recent", web.ActionHandler(getRecentModerationLog))
	router.GET("/api/moderation.log/pages/:count/:offset", web.ActionHandler(getModerationLogByCountAndOffsetAction))

	//session endpoints
	router.GET("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(getSessionKeyAction)))
	router.POST("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(setSessionKeyAction)))

	//auth endpoints
	router.GET("/oauth", web.ActionHandler(auth.SessionAwareAction(oauthAction)))
	router.GET("/logout", web.ActionHandler(auth.SessionRequiredAction(logoutAction)))
	router.POST("/logout", web.ActionHandler(auth.SessionRequiredAction(logoutAction)))

	//static files
	router.ServeFiles("/static/*filepath", http.Dir("server/_static"))

	router.NotFound = web.NotFoundHandler
	router.PanicHandler = web.PanicHandler

	return router
}

// Start starts the app.
func Start(router *httprouter.Router) {
	bindAddr := fmt.Sprintf(":%s", core.ConfigPort())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	web.Logf("Giffy Server Started, listening on %s", bindAddr)
	log.Fatal(server.ListenAndServe())
}
