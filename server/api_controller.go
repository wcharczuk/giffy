package server

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// APIController is the controller for api endpoints.
type APIController struct{}

func (api APIController) searchUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	users, err := model.SearchUsers(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(users)
}

func (api APIController) searchImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchImages(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func (api APIController) searchTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	query := ctx.Param("query")
	results, err := model.SearchTags(query, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(results)
}

func (api APIController) updateUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getRecentModerationLog(ctx *web.HTTPContext) web.ControllerResult {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(moderationLog)
}

func (api APIController) getModerationForUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	actions, err := model.GetModerationForUserID(user.ID, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(actions)
}

func (api APIController) getImageAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	images, err := model.GetAllImages(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(images)
}

func (api APIController) getRandomImagesAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")

	images, err := model.GetRandomImages(count, nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(images)
}

func (api APIController) getImagesForTagAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getTagsForImageAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getTagAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getTagsAction(ctx *web.HTTPContext) web.ControllerResult {
	tags, err := model.GetAllTags(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(tags)
}

func (api APIController) getUsersAction(ctx *web.HTTPContext) web.ControllerResult {
	users, err := model.GetAllUsers(nil)
	if err != nil {
		return ctx.InternalError(err)
	}
	return ctx.JSON(users)
}

func (api APIController) getUserAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getUserImagesAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) createImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	image, err := CreateImageFromFile(session.UserID, postedFile)
	if err != nil {
		return ctx.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)

	return ctx.JSON(image)
}

func (api APIController) createTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	tag := &model.Tag{}
	err := ctx.PostBodyAsJSON(tag)
	if err != nil {
		return ctx.BadRequest(err.Error())
	}

	if len(tag.TagValue) == 0 {
		return ctx.BadRequest("`tag_value` must be set.")
	}

	if len(strings.Trim(tag.TagValue, " \t")) == 0 {
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

func (api APIController) deleteImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) deleteTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getLinksForImageAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getLinksForTagAction(ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getVotesForUserForImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) getVotesForUserForTagAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) upvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return api.voteAction(true, session, ctx)
}

func (api APIController) downvoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return api.voteAction(false, session, ctx)
}

func (api APIController) voteAction(isUpvote bool, session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	didCreate, err := model.CreateOrIncrementVote(userID, image.ID, tag.ID, isUpvote, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	if didCreate {
		model.QueueModerationEntry(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID)
	}

	return ctx.OK()
}

func (api APIController) deleteUserVoteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

func (api APIController) deleteLinkAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID)

	return ctx.OK()
}

func (api APIController) getCurrentUserAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut(ctx)
		return ctx.JSON(cu)
	}
	cu.SetFromUser(session.User)
	return ctx.JSON(cu)
}

func (api APIController) getSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	value, hasValue := session.State[key]
	if !hasValue {
		return ctx.NotFound()
	}
	return ctx.JSON(value)
}

func (api APIController) setSessionKeyAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	key := ctx.RouteParameter("key")
	session.State[key] = ctx.PostBodyAsString()
	return ctx.OK()
}

func (api APIController) getModerationLogByCountAndOffsetAction(ctx *web.HTTPContext) web.ControllerResult {
	count := ctx.RouteParameterInt("count")
	offset := ctx.RouteParameterInt("offset")

	log, err := model.GetModerationLogByCountAndOffset(count, offset, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	return ctx.JSON(log)
}

// Register adds the routes to the router.
func (api APIController) Register(router *httprouter.Router) {
	router.GET("/api/users", web.ActionHandler(api.getUsersAction))
	router.GET("/api/user/:user_id", web.ActionHandler(api.getUserAction))
	router.PUT("/api/user/:user_id", web.ActionHandler(auth.SessionRequiredAction(api.updateUserAction)))

	router.GET("/api/user.images/:user_id", web.ActionHandler(api.getUserImagesAction))

	router.GET("/api/user.current", web.ActionHandler(auth.SessionAwareAction(api.getCurrentUserAction)))

	router.GET("/api/images", web.ActionHandler(api.getImagesAction))
	router.POST("/api/images", web.ActionHandler(auth.SessionRequiredAction(api.createImageAction)))
	router.GET("/api/images/random/:count", web.ActionHandler(api.getRandomImagesAction))

	router.GET("/api/image/:image_id", web.ActionHandler(api.getImageAction))
	router.DELETE("/api/image/:image_id", web.ActionHandler(auth.SessionRequiredAction(api.deleteImageAction)))

	router.GET("/api/tag.images/:tag_id", web.ActionHandler(api.getImagesForTagAction))
	router.GET("/api/image.tags/:image_id", web.ActionHandler(api.getTagsForImageAction))

	router.GET("/api/tags", web.ActionHandler(api.getTagsAction))
	router.POST("/api/tags", web.ActionHandler(auth.SessionRequiredAction(api.createTagAction)))
	router.GET("/api/tag/:tag_id", web.ActionHandler(api.getTagAction))
	router.DELETE("/api/tag/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.deleteTagAction)))

	router.GET("/api/image.votes/:image_id", web.ActionHandler(api.getLinksForImageAction))
	router.GET("/api/tag.votes/:tag_id", web.ActionHandler(api.getLinksForTagAction))

	router.DELETE("/api/link/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.deleteLinkAction)))

	router.GET("/api/user.votes.image/:image_id", web.ActionHandler(auth.SessionRequiredAction(api.getVotesForUserForImageAction)))
	router.GET("/api/user.votes.tag/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.getVotesForUserForTagAction)))
	router.DELETE("/api/user.vote/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.deleteUserVoteAction)))

	router.POST("/api/vote.up/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.upvoteAction)))
	router.POST("/api/vote.down/:image_id/:tag_id", web.ActionHandler(auth.SessionRequiredAction(api.downvoteAction)))

	router.GET("/api/users.search", web.ActionHandler(api.searchUsersAction))
	router.GET("/api/images.search", web.ActionHandler(api.searchImagesAction))
	router.GET("/api/tags.search", web.ActionHandler(api.searchTagsAction))

	router.GET("/api/moderation.log/recent", web.ActionHandler(api.getRecentModerationLog))
	router.GET("/api/moderation.log/pages/:count/:offset", web.ActionHandler(api.getModerationLogByCountAndOffsetAction))

	//session endpoints
	router.GET("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(api.getSessionKeyAction)))
	router.POST("/api/session/:key", web.ActionHandler(auth.SessionRequiredAction(api.setSessionKeyAction)))

}
