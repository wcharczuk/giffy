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

// GET "/api/users"
func (api API) getUsersAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)
	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	users, err := model.GetAllUsers(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(users)
}

// GET "/api/users.search"
func (api API) searchUsersAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	query := r.Param("query")
	users, err := model.SearchUsers(query, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(users)
}

// GET "/api/users/pages/:count/:offset"
func (api API) getUsersByCountAndOffsetAction(r *web.RequestContext) web.ControllerResult {
	if !auth.GetSession(r).User.IsAdmin {
		return r.API().NotAuthorized()
	}

	count := r.RouteParameterInt("count")
	offset := r.RouteParameterInt("offset")

	users, err := model.GetUsersByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(users)
}

// GET "/api/user/:user_id"
func (api API) getUserAction(r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if user.IsZero() {
		return r.API().NotFound()
	}

	return r.API().JSON(user)
}

// PUT "/api/user/:user_id"
func (api API) updateUserAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	userUUID := r.RouteParameter("user_id")

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	user, err := model.GetUserByUUID(userUUID, r.Tx())
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

// GET "/api/user.images/:user_id"
func (api API) getUserImagesAction(r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if user.IsZero() {
		return r.API().NotFound()
	}

	images, err := model.GetImagesForUserID(user.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(images)
}

// GET "/api/user.moderation/:user_id"
func (api API) getModerationForUserAction(r *web.RequestContext) web.ControllerResult {
	userUUID := r.RouteParameter("user_id")

	user, err := model.GetUserByUUID(userUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if user.IsZero() {
		return r.API().NotFound()
	}

	actions, err := model.GetModerationForUserID(user.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(actions)
}

// GET "/api/user.votes.image/:image_id"
func (api API) getVotesForUserForImageAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	votes, err := model.GetVotesForUserForImage(session.UserID, image.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(votes)
}

// GET "/api/user.votes.tag/:tag_id"
func (api API) getVotesForUserForTagAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	votes, err := model.GetVotesForUserForTag(session.UserID, tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(votes)
}

// DELETE "/api/user.vote/:image_id/:tag_id"
func (api API) deleteUserVoteAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")
	userID := session.UserID

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
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

	err = model.DeleteVote(userID, image.ID, tag.ID, r.Tx())
	if err != nil {
		tx.Rollback()
		return r.API().InternalError(err)
	}

	tx.Commit()
	return r.API().OK()
}

// GET "/api/images"
func (api API) getImagesAction(r *web.RequestContext) web.ControllerResult {
	images, err := model.GetAllImages(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(images)
}

// POST "/api/images"
func (api API) createImageAction(r *web.RequestContext) web.ControllerResult {
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

	session := auth.GetSession(r)
	image, err := CreateImageFromFile(session.UserID, !session.User.IsAdmin, postedFile.Contents, postedFile.Filename)
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
	return r.API().JSON(image)
}

// GET "/api/images/random/:count"
func (api API) getRandomImagesAction(r *web.RequestContext) web.ControllerResult {
	count := r.RouteParameterInt("count")

	images, err := model.GetRandomImages(count, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(images)
}

// GET "/api/images.search?query=<query>"
func (api API) searchImagesAction(r *web.RequestContext) web.ControllerResult {
	query := r.Param("query")
	results, err := model.SearchImages(query, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

// GET "/api/images.search/random/:count?query=<query>"
func (api API) searchImagesRandomAction(r *web.RequestContext) web.ControllerResult {
	count := r.RouteParameterInt("count")

	query := r.Param("query")
	results, err := model.SearchImagesRandom(query, count, r.Tx())

	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(results)
}

// GET "/api/image/:image_id"
func (api API) getImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	return r.API().JSON(image)
}

// PUT "/api/image/:image_id"
func (api API) updateImageAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	imageUUID := r.RouteParameter("image_id")

	if !session.User.IsModerator {
		return r.API().NotAuthorized()
	}

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
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

// DELETE "/api/image/:image_id"
func (api API) deleteImageAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID := r.RouteParameter("image_id")

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
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

	err = model.DeleteImageByID(image.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectImage, image.UUID)
	return r.API().OK()
}

// GET "/api/image.votes/:image_id"
func (api API) getLinksForImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForImage(image.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(voteSummaries)
}

// GET "/api/image.tags/:image_id"
func (api API) getTagsForImageAction(r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	results, err := model.GetTagsForImageID(image.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

// GET "/api/tags"
func (api API) getTagsAction(r *web.RequestContext) web.ControllerResult {
	tags, err := model.GetAllTags(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(tags)
}

// POST "/api/tags"
func (api API) createTagsAction(r *web.RequestContext) web.ControllerResult {
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

	session := auth.GetSession(r)

	tags := []*model.Tag{}
	for _, tagValue := range tagValues {
		existingTag, err := model.GetTagByValue(tagValue, r.Tx())

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

// GET "/api/tags.search?query=<query>"
func (api API) searchTagsAction(r *web.RequestContext) web.ControllerResult {
	query := r.Param("query")
	results, err := model.SearchTags(query, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

// GET "/api/tag/:tag_id"
func (api API) getTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if tag.IsZero() {
		tag, err = model.GetTagByValue(tagUUID, r.Tx())
		if err != nil {
			return r.API().InternalError(err)
		}
		if tag.IsZero() {
			return r.API().NotFound()
		}
	}

	return r.API().JSON(tag)
}

// DELETE "/api/tag/:tag_id"
func (api API) deleteTagAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	tagUUID := r.RouteParameter("tag_id")

	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return r.API().NotAuthorized()
	}

	err = model.DeleteTagWithVotesByID(tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID)
	return r.API().OK()
}

// GET "/api/tag.images/:tag_id"
func (api API) getImagesForTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}

	results, err := model.GetImagesForTagID(tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(results)
}

// GET "/api/tag.votes/:tag_id"
func (api API) getLinksForTagAction(r *web.RequestContext) web.ControllerResult {
	tagUUID := r.RouteParameter("tag_id")
	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForTag(tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().JSON(voteSummaries)
}

// DELETE "/api/link/:image_id/:tag_id"
func (api API) deleteLinkAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return r.API().NotAuthorized()
	}

	err = model.DeleteVoteSummary(image.ID, tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID)

	return r.API().OK()
}

// POST "/api/vote.up/:image_id/:tag_id"
func (api API) upvoteAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	return api.voteAction(true, session, r)
}

// POST "/api/vote.down/:image_id/:tag_id"
func (api API) downvoteAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	return api.voteAction(false, session, r)
}

func (api API) voteAction(isUpvote bool, session *auth.Session, r *web.RequestContext) web.ControllerResult {
	imageUUID := r.RouteParameter("image_id")
	tagUUID := r.RouteParameter("tag_id")
	userID := session.UserID

	tag, err := model.GetTagByUUID(tagUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if tag.IsZero() {
		return r.API().NotFound()
	}

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	existingUserVote, err := model.GetVote(userID, image.ID, tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return r.API().OK()
	}

	didCreate, err := model.CreateOrUpdateVote(userID, image.ID, tag.ID, isUpvote, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if didCreate {
		model.QueueModerationEntry(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID)
	}

	return r.API().OK()
}

// GET "/api/moderation.log/recent"
func (api API) getRecentModerationLogAction(r *web.RequestContext) web.ControllerResult {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(moderationLog)
}

// GET "/api/moderation.log/pages/:count/:offset"
func (api API) getModerationLogByCountAndOffsetAction(r *web.RequestContext) web.ControllerResult {
	count := r.RouteParameterInt("count")
	offset := r.RouteParameterInt("offset")

	log, err := model.GetModerationLogByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(log)
}

// GET "/api/search.history/recent"
func (api API) getRecentSearchHistoryAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	searchHistory, err := model.GetSearchHistory(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(searchHistory)
}

// GET "/api/search.history/pages/:count/:offset"
func (api API) getSearchHistoryByCountAndOffsetAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}

	count := r.RouteParameterInt("count")
	offset := r.RouteParameterInt("offset")

	searchHistory, err := model.GetSearchHistoryByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(searchHistory)
}

// GET "/api/stats"
func (api API) getSiteStatsAction(r *web.RequestContext) web.ControllerResult {
	stats, err := viewmodel.GetSiteStats(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().JSON(stats)
}

// GET "/api/session.user"
func (api API) getCurrentUserAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut()
		return r.API().JSON(cu)
	}
	cu.SetFromUser(session.User)
	return r.API().JSON(cu)
}

// GET "/api/session/:key"
func (api API) getSessionKeyAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	key := r.RouteParameter("key")
	value, hasValue := session.State[key]
	if !hasValue {
		return r.API().NotFound()
	}
	return r.API().JSON(value)
}

// POST "/api/session/:key"
func (api API) setSessionKeyAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	key := r.RouteParameter("key")
	session.State[key] = r.PostBodyAsString()
	return r.API().OK()
}

// GET "/api/jobs"
func (api API) getJobsStatusAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if !session.User.IsAdmin {
		return r.API().NotAuthorized()
	}
	status := chronometer.Default().Status()
	return r.API().JSON(status)
}

// POST "/api/job/:job_id"
func (api API) runJobAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

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

// POST "/api/logout"
func (api API) logoutAction(r *web.RequestContext) web.ControllerResult {
	session := auth.GetSession(r)

	if session == nil {
		return r.API().OK()
	}
	err := auth.Logout(session.UserID, session.SessionID, r, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().OK()
}

// Register adds the routes to the app.
func (api API) Register(app *web.App) {
	app.GET("/api/users", api.getUsersAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/users.search", api.searchUsersAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/users/pages/:count/:offset", api.getUsersByCountAndOffsetAction, auth.SessionRequired, web.InjectAPIProvider)

	app.GET("/api/user/:user_id", api.getUserAction)
	app.PUT("/api/user/:user_id", api.updateUserAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/user.images/:user_id", api.getUserImagesAction)
	app.GET("/api/user.moderation/:user_id", api.getModerationForUserAction)
	app.GET("/api/user.votes.image/:image_id", api.getVotesForUserForImageAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/user.votes.tag/:tag_id", api.getVotesForUserForTagAction, auth.SessionRequired, web.InjectAPIProvider)

	app.DELETE("/api/user.vote/:image_id/:tag_id", api.deleteUserVoteAction, auth.SessionRequired, web.InjectAPIProvider)

	app.GET("/api/images", api.getImagesAction)
	app.POST("/api/images", api.createImageAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/images/random/:count", api.getRandomImagesAction)
	app.GET("/api/images.search", api.searchImagesAction)
	app.GET("/api/images.search/random/:count", api.searchImagesRandomAction)

	app.GET("/api/image/:image_id", api.getImageAction)
	app.PUT("/api/image/:image_id", api.updateImageAction, auth.SessionRequired, web.InjectAPIProvider)
	app.DELETE("/api/image/:image_id", api.deleteImageAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/image.votes/:image_id", api.getLinksForImageAction)
	app.GET("/api/image.tags/:image_id", api.getTagsForImageAction)

	app.GET("/api/tags", api.getTagsAction)
	app.POST("/api/tags", api.createTagsAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/tags.search", api.searchTagsAction)

	app.GET("/api/tag/:tag_id", api.getTagAction)
	app.DELETE("/api/tag/:tag_id", api.deleteTagAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/tag.images/:tag_id", api.getImagesForTagAction)
	app.GET("/api/tag.votes/:tag_id", api.getLinksForTagAction)

	app.DELETE("/api/link/:image_id/:tag_id", api.deleteLinkAction, auth.SessionRequired, web.InjectAPIProvider)

	app.POST("/api/vote.up/:image_id/:tag_id", api.upvoteAction, auth.SessionRequired, web.InjectAPIProvider)
	app.POST("/api/vote.down/:image_id/:tag_id", api.downvoteAction, auth.SessionRequired, web.InjectAPIProvider)

	app.GET("/api/moderation.log/recent", api.getRecentModerationLogAction)
	app.GET("/api/moderation.log/pages/:count/:offset", api.getModerationLogByCountAndOffsetAction)

	app.GET("/api/search.history/recent", api.getRecentSearchHistoryAction, auth.SessionRequired, web.InjectAPIProvider)
	app.GET("/api/search.history/pages/:count/:offset", api.getSearchHistoryByCountAndOffsetAction, auth.SessionRequired, web.InjectAPIProvider)

	app.GET("/api/stats", api.getSiteStatsAction)

	//session endpoints
	app.GET("/api/session.user", api.getCurrentUserAction, auth.SessionAware, web.InjectAPIProvider)
	app.GET("/api/session/:key", api.getSessionKeyAction, auth.SessionRequired, web.InjectAPIProvider)
	app.POST("/api/session/:key", api.setSessionKeyAction, auth.SessionRequired, web.InjectAPIProvider)

	//jobs
	app.GET("/api/jobs", api.getJobsStatusAction, auth.SessionRequired, web.InjectAPIProvider)
	app.POST("/api/job/:job_id", api.runJobAction, auth.SessionRequired, web.InjectAPIProvider)

	// auth endpoints
	app.POST("/api/logout", api.logoutAction, auth.SessionRequired, web.InjectAPIProvider)
}
