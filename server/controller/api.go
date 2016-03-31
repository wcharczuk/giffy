package controller

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// API is the controller for api endpoints.
type API struct{}

func (api API) searchUsersAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	query := r.Param("query")
	users, err := model.SearchUsers(query, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(users)
}

func (api API) searchImagesAction(r *web.RequestContext) web.ControllerResult {
	query := r.Param("query")
	results, err := model.SearchImages(query, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
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
	ImageUUID    string        `json:"image_uuid"`
	ResponseType string        `json:"response_type"`
	Text         string        `json:"text,omitempty"`
	Attachments  []interface{} `json:"attachments"`
}

func (api API) searchTagsAction(r *web.RequestContext) web.ControllerResult {
	query := r.Param("query")
	results, err := model.SearchTags(query, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

func (api API) updateUserAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if user.IsZero() {
		return r.API().NotFound()
	}

	var postedUser model.User
	err = r.PostBodyAsJSON(&postedUser)
	if err != nil {
		return r.API().BadRequest("Post body was not properly formatted.")
	}

	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	postedUser.CreatedUTC = user.CreatedUTC
	postedUser.Username = user.Username

	if !user.IsAdmin && postedUser.IsAdmin {
		return r.API().BadRequest("Cannot promote user to admin through the UI; this must be done in the db directly.")
	}

	if user.IsAdmin && !postedUser.IsAdmin {
		return r.API().BadRequest("Cannot demote user from admin through the UI; this must be done in the db directly.")
	}

	if postedUser.IsAdmin && postedUser.IsBanned {
		return r.API().BadRequest("Cannot ban admins.")
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
		return r.API().InternalError(err)
	}

	return r.API().JSON(postedUser)
}

func (api API) getRecentModerationLog(r *web.RequestContext) web.ControllerResult {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(moderationLog)
}

func (api API) getModerationForUserAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	if !session.User.IsModerator {
		return r.API().NotAuthorized()
	}

	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if user.IsZero() {
		return r.API().NotFound()
	}

	actions, err := model.GetModerationForUserID(user.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(actions)
}

func (api API) getImagesAction(r *web.RequestContext) web.ControllerResult {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(images)
}

func (api API) getRandomImagesAction(r *web.RequestContext) web.ControllerResult {
	count := r.RouteParameterInt("count")

	images, err := model.GetRandomImages(count, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(images)
}

func (api API) getImagesForTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}

	results, err := model.GetImagesForTagID(tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

func (api API) getImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	return r.API().JSON(image)
}

func (api API) updateImageAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")

	if !session.User.IsModerator {
		return r.API().NotAuthorized()
	}

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	updatedImage := model.Image{}
	err = r.PostBodyAsJSON(&updatedImage)

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
		return r.API().InternalError(err)
	}

	return r.API().JSON(image)
}

func (api API) getTagsForImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	results, err := model.GetTagsForImageID(image.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

func (api API) getTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if tag.IsZero() {
		tag, err = model.GetTagByValue(tagUUID, nil)
		if err != nil {
			return r.API().InternalError(err)
		}
		if tag.IsZero() {
			return r.API().NotFound()
		}
	}

	return r.API().JSON(tag)
}

func (api API) getTagsAction(r *web.RequestContext) web.ControllerResult {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(tags)
}

func (api API) getUsersAction(r *web.RequestContext) web.ControllerResult {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(users)
}

func (api API) getUserAction(r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if user.IsZero() {
		return r.API().NotFound()
	}

	return r.API().JSON(user)
}

func (api API) getUserImagesAction(r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if user.IsZero() {
		return r.API().NotFound()
	}

	images, err := model.GetImagesForUserID(user.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(images)
}

func (api API) createImageAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	files, filesErr := r.PostedFiles()
	if filesErr != nil {
		return r.API().BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return r.API().BadRequest("No files posted.")
	}

	postedFile := files[0]
	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if !existing.IsZero() {
		return r.API().JSON(existing)
	}

	image, err := CreateImageFromFile(session.UserID, postedFile.Contents, postedFile.Filename)
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
	return r.API().JSON(image)
}

func (api API) createTagAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	args := viewmodel.CreateTagArgs{}
	err := r.PostBodyAsJSON(&args)
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	if len(args) == 0 {
		return r.API().BadRequest("empty post body, please submit an array of strings.")
	}

	var tagValues []string
	for _, value := range args {
		tagValue := model.CleanTagValue(value)
		if len(tagValue) == 0 {
			return r.API().BadRequest("`tag_value` be in the form [a-z,A-Z,0-9]+")
		}

		tagValues = append(tagValues, tagValue)
	}

	tags := []*model.Tag{}
	for _, tagValue := range tagValues {
		existingTag, err := model.GetTagByValue(tagValue, nil)

		if err != nil {
			return r.API().InternalError(err)
		}

		if !existingTag.IsZero() {
			tags = append(tags, existingTag)
			continue
		}

		tag := model.NewTag(session.UserID, tagValue)
		err = spiffy.DefaultDb().Create(tag)
		if err != nil {
			return r.API().InternalError(err)
		}
		model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID)
		tags = append(tags, tag)
	}

	return r.API().JSON(tags)
}

func (api API) deleteImageAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID := r.RouteParameter("image_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	if !currentUser.IsModerator && image.CreatedBy != currentUser.ID {
		return r.API().NotAuthorized()
	}

	//delete from s3 (!!)
	err = filecache.DeleteFile(filecache.NewLocationFromKey(image.S3Key))
	if err != nil {
		return r.API().InternalError(err)
	}

	err = model.DeleteImageByID(image.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectImage, image.UUID)
	return r.API().OK()
}

func (api API) deleteTagAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	tagUUID := r.RouteParameter("tag_id")

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return r.API().NotAuthorized()
	}

	err = model.DeleteTagWithVotesByID(tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID)
	return r.API().OK()
}

func (api API) getLinksForImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForImage(image.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(voteSummaries)
}

func (api API) getLinksForTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForTag(tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(voteSummaries)
}

func (api API) getVotesForUserForImageAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	votes, err := model.GetVotesForUserForImage(session.UserID, image.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(votes)
}

func (api API) getVotesForUserForTagAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	votes, err := model.GetVotesForUserForTag(session.UserID, tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(votes)
}

func (api API) upvoteAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	return api.voteAction(true, session, r)
}

func (api API) downvoteAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	return api.voteAction(false, session, r)
}

func (api API) voteAction(isUpvote bool, session *auth.Session, r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")
	userID := session.UserID

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	existingUserVote, err := model.GetVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return r.API().OK()
	}

	didCreate, err := model.CreateOrUpdateVote(userID, image.ID, tag.ID, isUpvote, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	if didCreate {
		model.QueueModerationEntry(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID)
	}

	return r.API().OK()
}

func (api API) deleteUserVoteAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")
	userID := session.UserID

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}

	tx, err := spiffy.DefaultDb().Begin()

	vote, err := model.GetVote(userID, image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return r.API().InternalError(err)
	}
	if vote.IsZero() {
		tx.Rollback()
		return r.API().NotFound()
	}

	// was it an upvote or downvote
	wasUpvote := vote.IsUpvote

	// adjust the vote summary ...
	voteSummary, err := model.GetVoteSummary(image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return r.API().InternalError(err)
	}

	if wasUpvote {
		voteSummary.VotesFor--
	} else {
		voteSummary.VotesAgainst--
	}

	err = model.SetVoteSummaryVoteCounts(image.ID, tag.ID, voteSummary.VotesFor, voteSummary.VotesAgainst, tx)
	if err != nil {
		tx.Rollback()
		return r.API().InternalError(err)
	}

	err = model.DeleteVote(userID, image.ID, tag.ID, nil)
	if err != nil {
		tx.Rollback()
		return r.API().InternalError(err)
	}

	tx.Commit()
	return r.API().OK()
}

func (api API) deleteLinkAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	currentUser, err := model.GetUserByID(session.UserID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")

	image, err := model.GetImageByUUID(imageUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return r.API().NotAuthorized()
	}

	err = model.DeleteVoteSummary(image.ID, tag.ID, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID)

	return r.API().OK()
}

func (api API) getCurrentUserAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut()
		return r.API().JSON(cu)
	}
	cu.SetFromUser(session.User)
	return r.API().JSON(cu)
}

func (api API) getSessionKeyAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	key := r.RouteParameter("key")
	value, hasValue := session.State[key]
	if !hasValue {
		return r.API().NotFound()
	}
	return r.API().JSON(value)
}

func (api API) setSessionKeyAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	key := r.RouteParameter("key")
	session.State[key] = r.PostBodyAsString()
	return r.API().OK()
}

func (api API) getModerationLogByCountAndOffsetAction(r *web.RequestContext) web.ControllerResult {
	count := r.RouteParameterInt("count")
	offset := r.RouteParameterInt("offset")

	log, err := model.GetModerationLogByCountAndOffset(count, offset, nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(log)
}

func (api API) logoutAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	if session == nil {
		return r.API().OK()
	}
	err := auth.Logout(session.UserID, session.SessionID)
	if err != nil {
		return r.API().InternalError(err)
	}
	r.ExpireCookie(auth.SessionParamName)

	return r.API().OK()
}

func (api API) getSiteStatsAction(r *web.RequestContext) web.ControllerResult {
	stats, err := viewmodel.GetSiteStats(nil)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(stats)
}

func (api API) getJobsStatusAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}
	status := chronometer.Default().Status()
	return r.API().JSON(status)
}

func (api API) runJobAction(session *auth.Session, r *web.RequestContext) web.ControllerResult {
	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}
	jobID := r.RouteParameter("job_id")

	err := chronometer.Default().RunJob(jobID)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().OK()
}

// Register adds the routes to the app.
func (api API) Register(app *web.App) {
	app.GET("/api/users", api.getUsersAction)
	app.GET("/api/users.search", auth.SessionRequiredAction(web.ProviderAPI, api.searchUsersAction))

	app.GET("/api/user/:user_id", api.getUserAction)
	app.PUT("/api/user/:user_id", auth.SessionRequiredAction(web.ProviderAPI, api.updateUserAction))
	app.GET("/api/user.images/:user_id", api.getUserImagesAction)
	app.GET("/api/user.votes.image/:image_id", auth.SessionRequiredAction(web.ProviderAPI, api.getVotesForUserForImageAction))
	app.GET("/api/user.votes.tag/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.getVotesForUserForTagAction))
	app.DELETE("/api/user.vote/:image_id/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.deleteUserVoteAction))

	app.GET("/api/images", api.getImagesAction)
	app.POST("/api/images", auth.SessionRequiredAction(web.ProviderAPI, api.createImageAction))
	app.GET("/api/images/random/:count", api.getRandomImagesAction)
	app.GET("/api/images.search", api.searchImagesAction)

	app.GET("/api/image/:image_id", api.getImageAction)
	app.PUT("/api/image/:image_id", auth.SessionRequiredAction(web.ProviderAPI, api.updateImageAction))
	app.DELETE("/api/image/:image_id", auth.SessionRequiredAction(web.ProviderAPI, api.deleteImageAction))
	app.GET("/api/image.votes/:image_id", api.getLinksForImageAction)
	app.GET("/api/image.tags/:image_id", api.getTagsForImageAction)

	app.GET("/api/tags", api.getTagsAction)
	app.POST("/api/tags", auth.SessionRequiredAction(web.ProviderAPI, api.createTagAction))
	app.GET("/api/tags.search", api.searchTagsAction)

	app.GET("/api/tag/:tag_id", api.getTagAction)
	app.DELETE("/api/tag/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.deleteTagAction))
	app.GET("/api/tag.images/:tag_id", api.getImagesForTagAction)
	app.GET("/api/tag.votes/:tag_id", api.getLinksForTagAction)

	app.DELETE("/api/link/:image_id/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.deleteLinkAction))

	app.POST("/api/vote.up/:image_id/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.upvoteAction))
	app.POST("/api/vote.down/:image_id/:tag_id", auth.SessionRequiredAction(web.ProviderAPI, api.downvoteAction))

	app.GET("/api/moderation.log/recent", api.getRecentModerationLog)
	app.GET("/api/moderation.log/pages/:count/:offset", api.getModerationLogByCountAndOffsetAction)

	app.GET("/api/stats", api.getSiteStatsAction)

	//session endpoints
	app.GET("/api/session.user", auth.SessionAwareAction(web.ProviderAPI, api.getCurrentUserAction))
	app.GET("/api/session/:key", auth.SessionRequiredAction(web.ProviderAPI, api.getSessionKeyAction))
	app.POST("/api/session/:key", auth.SessionRequiredAction(web.ProviderAPI, api.setSessionKeyAction))

	//jobs
	app.GET("/api/jobs", auth.SessionRequiredAction(web.ProviderAPI, api.getJobsStatusAction))
	app.POST("/api/job/:job_id", auth.SessionRequiredAction(web.ProviderAPI, api.runJobAction))

	// auth endpoints
	app.POST("/api/logout", auth.SessionAwareAction(web.ProviderAPI, api.logoutAction))
}
