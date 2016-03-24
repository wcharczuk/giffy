package controller

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// API is the controller for api endpoints.
type API struct{}

func (api API) searchUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	users, err := model.SearchUsers(query, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(users)
}

func (api API) searchImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchImages(query, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(results)
}

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackMessageAttachment struct {
	Text   string       `json:"text"`
	Fields []slackField `json:"field"`
}

type slackImageAttachment struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

type slackResponse struct {
	ResponseType string        `json:"response_type"`
	Text         string        `json:"text,omitempty"`
	Attachments  []interface{} `json:"attachments"`
}

func (api API) searchImagesSlackAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("text")

	var result *model.Image
	var err error
	if strings.HasPrefix(query, "img:") {
		uuid := strings.Replace(query, "img:", "", -1)
		result, err = model.GetImageByUUID(uuid, nil)
	} else {
		result, err = model.SearchImagesSlack(query, nil)
	}
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if result.IsZero() {
		return ctx.Raw("text/plaid; charset=utf-8", []byte(fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())))
	}

	res := slackResponse{}
	res.ResponseType = "in_channel"

	if !strings.HasPrefix(query, "img:") {
		res.Attachments = []interface{}{
			slackImageAttachment{Title: query, ImageURL: result.S3ReadURL},
		}
	} else {
		res.Attachments = []interface{}{
			slackImageAttachment{Title: result.Tags[0].TagValue, ImageURL: result.S3ReadURL},
		}
	}

	responseBytes, err := json.Marshal(res)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.Raw("application/json; charset=utf-8", responseBytes)
}

func (api API) searchTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchTags(query, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(results)
}

func (api API) updateUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	if !session.User.IsAdmin {
		return ctx.API.NotAuthorized()
	}

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if user.IsZero() {
		return ctx.API.NotFound()
	}

	var postedUser model.User
	err = ctx.PostBodyAsJSON(&postedUser)
	if err != nil {
		return ctx.API.BadRequest("Post body was not properly formatted.")
	}

	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	postedUser.CreatedUTC = user.CreatedUTC
	postedUser.Username = user.Username

	if !user.IsAdmin && postedUser.IsAdmin {
		return ctx.API.BadRequest("Cannot promote user to admin through the UI; this must be done in the db directly.")
	}

	if user.IsAdmin && !postedUser.IsAdmin {
		return ctx.API.BadRequest("Cannot demote user from admin through the UI; this must be done in the db directly.")
	}

	if postedUser.IsAdmin && postedUser.IsBanned {
		return ctx.API.BadRequest("Cannot ban admins.")
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
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(postedUser)
}

func (api API) getRecentModerationLog(ctx *web.HTTPContext) web.ControllerResult {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(moderationLog)
}

func (api API) getModerationForUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsModerator {
		return ctx.API.NotAuthorized()
	}

	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if user.IsZero() {
		return ctx.API.NotFound()
	}

	actions, err := model.GetModerationForUserID(user.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(actions)
}

func (api API) getImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(images)
}

func (api API) getRandomImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")

	images, err := model.GetRandomImages(count, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(images)
}

func (api API) getImagesForTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}

	results, err := model.GetImagesForTagID(tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(results)
}

func (api API) getImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}
	return ctx.API.JSON(image)
}

func (api API) updateImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")

	if !session.User.IsModerator {
		return ctx.API.NotAuthorized()
	}

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}

	updatedImage := model.Image{}
	err = ctx.PostBodyAsJSON(&updatedImage)

	if len(updatedImage.DisplayName) != 0 {
		image.DisplayName = updatedImage.DisplayName
	}

	if updatedImage.IsCensored != image.IsCensored {
		if updatedImage.IsCensored {
			model.QueueModerationEntry(session.UserID, model.ModerationVerbCensor, model.ModerationObjectImage, image.UUID)
		} else {
			model.QueueModerationEntry(session.UserID, model.ModerationVerbUncensor, model.ModerationObjectImage, image.UUID)
		}

		image.IsCensored = updatedImage.IsCensored
	}

	err = spiffy.DefaultDb().Update(image)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(image)
}

func (api API) getTagsForImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}

	results, err := model.GetTagsForImageID(image.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(results)
}

func (api API) getTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if tag.IsZero() {
		tag, err = model.GetTagByValue(tagUUID, nil)
		if err != nil {
			return ctx.API.InternalError(err)
		}
		if tag.IsZero() {
			return ctx.API.NotFound()
		}
	}

	return ctx.API.JSON(tag)
}

func (api API) getTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(tags)
}

func (api API) getUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(users)
}

func (api API) getUserAction(ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if user.IsZero() {
		return ctx.API.NotFound()
	}

	return ctx.API.JSON(user)
}

func (api API) getUserImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	userUUID := ctx.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if user.IsZero() {
		return ctx.API.NotFound()
	}

	images, err := model.GetImagesForUserID(user.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(images)
}

func (api API) createImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.API.BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return ctx.API.BadRequest("No files posted.")
	}

	postedFile := files[0]

	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if !existing.IsZero() {
		return ctx.API.JSON(existing)
	}

	image, err := CreateImageFromFile(session.UserID, postedFile)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)

	return ctx.API.JSON(image)
}

func (api API) createTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	tag := &model.Tag{}
	err := ctx.PostBodyAsJSON(tag)
	if err != nil {
		return ctx.API.BadRequest(err.Error())
	}

	if len(tag.TagValue) == 0 {
		return ctx.API.BadRequest("`tag_value` must be set.")
	}

	tagValue := model.CleanTagValue(tag.TagValue)
	if len(tagValue) == 0 {
		return ctx.API.BadRequest("`tag_value` must be set.")
	}

	//check if the tag exists first
	existingTag, err := model.GetTagByValue(tagValue, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if !existingTag.IsZero() {
		return ctx.API.JSON(existingTag)
	}

	tag.TagValue = tagValue
	tag.UUID = core.UUIDv4().ToShortString()
	tag.CreatedUTC = time.Now().UTC()
	tag.CreatedBy = session.UserID
	tag.TagValue = strings.ToLower(tag.TagValue)

	err = spiffy.DefaultDb().Create(tag)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID)
	return ctx.API.JSON(tag)
}

func (api API) deleteImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	imageUUID := ctx.RouteParameter("image_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}
	if !currentUser.IsModerator && image.CreatedBy != currentUser.ID {
		return ctx.API.NotAuthorized()
	}

	//delete from s3 (!!)
	err = filecache.DeleteFile(filecache.NewLocationFromKey(image.S3Key))
	if err != nil {
		return ctx.API.InternalError(err)
	}

	err = model.DeleteImageByID(image.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.OK()
}

func (api API) deleteTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	tagUUID := ctx.RouteParameter("tag_id")

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return ctx.API.NotAuthorized()
	}

	err = model.DeleteTagByID(tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID)
	return ctx.API.OK()
}

func (api API) getLinksForImageAction(ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForImage(image.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(voteSummaries)
}

func (api API) getLinksForTagAction(ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForTag(tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(voteSummaries)
}

func (api API) getVotesForUserForImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}
	votes, err := model.GetVotesForUserForImage(session.UserID, image.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(votes)
}

func (api API) getVotesForUserForTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	tagUUID := ctx.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}
	votes, err := model.GetVotesForUserForTag(session.UserID, tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.JSON(votes)
}

func (api API) upvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return api.voteAction(true, session, ctx)
}

func (api API) downvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return api.voteAction(false, session, ctx)
}

func (api API) voteAction(isUpvote bool, session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")
	userID := session.UserID

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}

	existingUserVote, err := model.GetVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return ctx.API.OK()
	}

	didCreate, err := model.CreateOrIncrementVote(userID, image.ID, tag.ID, isUpvote, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	if didCreate {
		model.QueueModerationEntry(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID)
	}

	return ctx.API.OK()
}

func (api API) deleteUserVoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")
	userID := session.UserID

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}

	tx, err := spiffy.DefaultDb().Begin()

	vote, err := model.GetVote(userID, image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return ctx.API.InternalError(err)
	}
	if vote.IsZero() {
		tx.Rollback()
		return ctx.API.NotFound()
	}

	// was it an upvote or downvote
	wasUpvote := vote.IsUpvote

	// adjust the vote summary ...
	voteSummary, err := model.GetVoteSummary(image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return ctx.API.InternalError(err)
	}

	if wasUpvote {
		voteSummary.VotesFor--
	} else {
		voteSummary.VotesAgainst--
	}

	err = model.SetVoteCount(image.ID, tag.ID, voteSummary.VotesFor, voteSummary.VotesAgainst, tx)
	if err != nil {
		tx.Rollback()
		return ctx.API.InternalError(err)
	}

	err = model.DeleteVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		tx.Rollback()
		return ctx.API.InternalError(err)
	}

	tx.Commit()
	return ctx.API.OK()
}

func (api API) deleteLinkAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	imageUUID := ctx.RouteParameter("image_id")
	tagUUID := ctx.RouteParameter("tag_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if image.IsZero() {
		return ctx.API.NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	if tag.IsZero() {
		return ctx.API.NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return ctx.API.NotAuthorized()
	}

	err = model.DeleteVoteSummary(image.ID, tag.ID, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID)

	return ctx.API.OK()
}

func (api API) getCurrentUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut(ctx)
		return ctx.API.JSON(cu)
	}
	cu.SetFromUser(session.User)
	return ctx.API.JSON(cu)
}

func (api API) getSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	value, hasValue := session.State[key]
	if !hasValue {
		return ctx.API.NotFound()
	}
	return ctx.API.JSON(value)
}

func (api API) setSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	session.State[key] = ctx.PostBodyAsString()
	return ctx.API.OK()
}

func (api API) getModerationLogByCountAndOffsetAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")
	offset := ctx.RouteParameterInt("offset")

	log, err := model.GetModerationLogByCountAndOffset(count, offset, nil)
	if err != nil {
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(log)
}

func (api API) logoutAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if session == nil {
		return ctx.API.OK()
	}
	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	ctx.ExpireCookie(auth.SessionParamName)

	return ctx.API.OK()
}

func (api API) getSiteStatsAction(ctx *web.HTTPContext) web.ControllerResult {
	stats, err := viewmodel.GetSiteStats()
	if err != nil {
		return ctx.API.InternalError(err)
	}

	return ctx.API.JSON(stats)
}

func (api API) getJobsStatusAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsAdmin {
		return ctx.API.NotAuthorized()
	}
	status := chronometer.Default().Status()
	return ctx.API.JSON(status)
}

func (api API) runJobAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsAdmin {
		return ctx.API.NotAuthorized()
	}
	jobID := ctx.RouteParameter("job_id")

	err := chronometer.Default().RunJob(jobID)
	if err != nil {
		return ctx.API.InternalError(err)
	}
	return ctx.API.OK()
}

// Register adds the routes to the router.
func (api API) Register(router *httprouter.Router) {
	router.GET("/api/users", web.ActionHandler(api.getUsersAction))
	router.GET("/api/user/:user_id", web.ActionHandler(api.getUserAction))
	router.PUT("/api/user/:user_id", auth.APISessionRequiredAction(api.updateUserAction))

	router.GET("/api/user.images/:user_id", web.ActionHandler(api.getUserImagesAction))

	router.GET("/api/images", web.ActionHandler(api.getImagesAction))
	router.POST("/api/images", auth.APISessionRequiredAction(api.createImageAction))
	router.GET("/api/images/random/:count", web.ActionHandler(api.getRandomImagesAction))

	router.GET("/api/image/:image_id", web.ActionHandler(api.getImageAction))
	router.PUT("/api/image/:image_id", auth.APISessionRequiredAction(api.updateImageAction))
	router.DELETE("/api/image/:image_id", auth.APISessionRequiredAction(api.deleteImageAction))

	router.GET("/api/tag.images/:tag_id", web.ActionHandler(api.getImagesForTagAction))
	router.GET("/api/image.tags/:image_id", web.ActionHandler(api.getTagsForImageAction))

	router.GET("/api/tags", web.ActionHandler(api.getTagsAction))
	router.POST("/api/tags", auth.APISessionRequiredAction(api.createTagAction))
	router.GET("/api/tag/:tag_id", web.ActionHandler(api.getTagAction))
	router.DELETE("/api/tag/:tag_id", auth.APISessionRequiredAction(api.deleteTagAction))

	router.GET("/api/image.votes/:image_id", web.ActionHandler(api.getLinksForImageAction))
	router.GET("/api/tag.votes/:tag_id", web.ActionHandler(api.getLinksForTagAction))

	router.DELETE("/api/link/:image_id/:tag_id", auth.APISessionRequiredAction(api.deleteLinkAction))

	router.GET("/api/user.votes.image/:image_id", auth.APISessionRequiredAction(api.getVotesForUserForImageAction))
	router.GET("/api/user.votes.tag/:tag_id", auth.APISessionRequiredAction(api.getVotesForUserForTagAction))
	router.DELETE("/api/user.vote/:image_id/:tag_id", auth.APISessionRequiredAction(api.deleteUserVoteAction))

	router.POST("/api/vote.up/:image_id/:tag_id", auth.APISessionRequiredAction(api.upvoteAction))
	router.POST("/api/vote.down/:image_id/:tag_id", auth.APISessionRequiredAction(api.downvoteAction))

	router.GET("/api/users.search", web.ActionHandler(api.searchUsersAction))
	router.GET("/api/images.search", web.ActionHandler(api.searchImagesAction))
	router.GET("/api/images.search/slack", web.ActionHandler(api.searchImagesSlackAction))
	router.POST("/api/images.search/slack", web.ActionHandler(api.searchImagesSlackAction))
	router.GET("/api/tags.search", web.ActionHandler(api.searchTagsAction))

	router.GET("/api/moderation.log/recent", web.ActionHandler(api.getRecentModerationLog))
	router.GET("/api/moderation.log/pages/:count/:offset", web.ActionHandler(api.getModerationLogByCountAndOffsetAction))

	router.GET("/api/stats", web.ActionHandler(api.getSiteStatsAction))

	//session endpoints
	router.GET("/api/session.user", auth.APISessionAwareAction(api.getCurrentUserAction))
	router.GET("/api/session/:key", auth.APISessionRequiredAction(api.getSessionKeyAction))
	router.POST("/api/session/:key", auth.APISessionRequiredAction(api.setSessionKeyAction))

	//jobs
	router.GET("/api/jobs", auth.APISessionRequiredAction(api.getJobsStatusAction))
	router.POST("/api/job/:job_id", auth.APISessionRequiredAction(api.runJobAction))

	// auth endpoints
	router.POST("/api/logout", auth.APISessionAwareAction(api.logoutAction))
}
