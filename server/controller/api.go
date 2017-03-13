package controller

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/filecache"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
	"github.com/wcharczuk/giffy/server/webutil"
)

// API is the controller for api endpoints.
type API struct{}

// Register adds the routes to the app.
func (api API) Register(app *web.App) {
	app.GET("/api/users", api.getUsersAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/users.search", api.searchUsersAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/users/pages/:count/:offset", api.getUsersByCountAndOffsetAction, web.SessionRequired, web.APIProviderAsDefault)

	app.GET("/api/user/:user_id", api.getUserAction)
	app.PUT("/api/user/:user_id", api.updateUserAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/user.images/:user_id", api.getUserImagesAction)
	app.GET("/api/user.moderation/:user_id", api.getModerationForUserAction)
	app.GET("/api/user.votes.image/:image_id", api.getVotesForUserForImageAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/user.votes.tag/:tag_id", api.getVotesForUserForTagAction, web.SessionRequired, web.APIProviderAsDefault)

	app.DELETE("/api/user.vote/:image_id/:tag_id", api.deleteUserVoteAction, web.SessionRequired, web.APIProviderAsDefault)

	app.GET("/api/images", api.getImagesAction)
	app.POST("/api/images", api.createImageAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/images/random/:count", api.getRandomImagesAction, web.SessionAware, web.APIProviderAsDefault)
	app.GET("/api/images.search", api.searchImagesAction, web.SessionAware, web.APIProviderAsDefault)
	app.GET("/api/images.search/random/:count", api.searchImagesRandomAction, web.SessionAware, web.APIProviderAsDefault)

	app.GET("/api/image/:image_id", api.getImageAction)
	app.PUT("/api/image/:image_id", api.updateImageAction, web.SessionRequired, web.APIProviderAsDefault)
	app.DELETE("/api/image/:image_id", api.deleteImageAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/image.votes/:image_id", api.getLinksForImageAction)
	app.GET("/api/image.tags/:image_id", api.getTagsForImageAction)

	app.GET("/api/tags", api.getTagsAction)
	app.POST("/api/tags", api.createTagsAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/tags/random/:count", api.getRandomTagsAction)
	app.GET("/api/tags.search", api.searchTagsAction)
	app.GET("/api/tags.search/random/:count", api.searchTagsRandomAction)

	app.GET("/api/tag/:tag_id", api.getTagAction)
	app.DELETE("/api/tag/:tag_id", api.deleteTagAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/tag.images/:tag_id", api.getImagesForTagAction)
	app.GET("/api/tag.votes/:tag_id", api.getLinksForTagAction)

	app.GET("/api/teams", api.getTeamsAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/team/:team_id", api.getTeamAction, web.SessionRequired, web.APIProviderAsDefault)
	app.POST("/api/team/:team_id", api.createTeamAction, web.SessionRequired, web.APIProviderAsDefault)
	app.PUT("/api/team/:team_id", api.updateTeamAction, web.SessionRequired, web.APIProviderAsDefault)
	app.PATCH("/api/team/:team_id", api.patchTeamAction, web.SessionRequired, web.APIProviderAsDefault)
	app.DELETE("/api/team/:team_id", api.deleteTeamAction, web.SessionRequired, web.APIProviderAsDefault)

	app.DELETE("/api/link/:image_id/:tag_id", api.deleteLinkAction, web.SessionRequired, web.APIProviderAsDefault)

	app.POST("/api/vote.up/:image_id/:tag_id", api.upvoteAction, web.SessionRequired, web.APIProviderAsDefault)
	app.POST("/api/vote.down/:image_id/:tag_id", api.downvoteAction, web.SessionRequired, web.APIProviderAsDefault)

	app.GET("/api/moderation.log/recent", api.getRecentModerationLogAction)
	app.GET("/api/moderation.log/pages/:count/:offset", api.getModerationLogByCountAndOffsetAction)

	app.GET("/api/search.history/recent", api.getRecentSearchHistoryAction, web.SessionRequired, web.APIProviderAsDefault)
	app.GET("/api/search.history/pages/:count/:offset", api.getSearchHistoryByCountAndOffsetAction, web.SessionRequired, web.APIProviderAsDefault)

	app.GET("/api/stats", api.getSiteStatsAction)
	app.GET("/api/image.stats/:image_id", api.getImageStatsAction)

	//session endpoints
	app.GET("/api/session.user", api.getCurrentUserAction, web.SessionAware, web.APIProviderAsDefault)
	app.GET("/api/session/:key", api.getSessionKeyAction, web.SessionRequired, web.APIProviderAsDefault)
	app.POST("/api/session/:key", api.setSessionKeyAction, web.SessionRequired, web.APIProviderAsDefault)
	app.PUT("/api/session/:key", api.setSessionKeyAction, web.SessionRequired, web.APIProviderAsDefault)
	app.DELETE("/api/session/:key", api.deleteSessionKeyAction, web.SessionRequired, web.APIProviderAsDefault)

	//jobs
	app.GET("/api/jobs", api.getJobsStatusAction, web.SessionRequired, web.APIProviderAsDefault)
	app.POST("/api/job/:job_id", api.runJobAction, web.SessionRequired, web.APIProviderAsDefault)

	//errors
	app.GET("/api/errors/:limit/:offset", api.getErrorsAction, web.SessionRequired, web.APIProviderAsDefault)

	// auth endpoints
	app.POST("/api/logout", api.logoutAction, web.SessionRequired, web.APIProviderAsDefault)
}

// GET "/api/users"
func (api API) getUsersAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return r.API().NotAuthorized()
	}

	users, err := model.GetAllUsers(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(users)
}

// GET "/api/users.search"
func (api API) searchUsersAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return r.API().NotAuthorized()
	}

	query := r.Param("query")
	users, err := model.SearchUsers(query, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(users)
}

// GET "/api/users/pages/:count/:offset"
func (api API) getUsersByCountAndOffsetAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return r.API().NotAuthorized()
	}

	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	users, err := model.GetUsersByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(users)
}

// GET "/api/user/:user_id"
func (api API) getUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	user, err := model.GetUserByUUID(userUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	if user.IsZero() {
		return r.API().NotFound()
	}

	return r.API().Result(user)
}

// PUT "/api/user/:user_id"
func (api API) updateUserAction(r *web.Ctx) web.Result {
	session := r.Session()
	sessionUser := webutil.GetUser(session)
	if !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
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

	var moderationEntry *model.Moderation
	if !user.IsModerator && postedUser.IsModerator {
		moderationEntry = model.NewModeration(session.UserID, model.ModerationVerbPromoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsModerator && !postedUser.IsModerator {
		moderationEntry = model.NewModeration(session.UserID, model.ModerationVerbDemoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	}

	if !user.IsBanned && postedUser.IsBanned {
		moderationEntry = model.NewModeration(session.UserID, model.ModerationVerbBan, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsBanned && !postedUser.IsBanned {
		moderationEntry = model.NewModeration(session.UserID, model.ModerationVerbUnban, model.ModerationObjectUser, postedUser.UUID)
	}

	r.Logger().OnEvent(core.EventFlagModeration, moderationEntry)

	err = spiffy.DB().Update(&postedUser)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(postedUser)
}

// GET "/api/user.images/:user_id"
func (api API) getUserImagesAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

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

	return r.API().Result(images)
}

// GET "/api/user.moderation/:user_id"
func (api API) getModerationForUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

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
	return r.API().Result(actions)
}

// GET "/api/user.votes.image/:image_id"
func (api API) getVotesForUserForImageAction(r *web.Ctx) web.Result {
	session := r.Session()

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(votes)
}

// GET "/api/user.votes.tag/:tag_id"
func (api API) getVotesForUserForTagAction(r *web.Ctx) web.Result {
	session := r.Session()

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(votes)
}

// DELETE "/api/user.vote/:image_id/:tag_id"
func (api API) deleteUserVoteAction(r *web.Ctx) web.Result {
	session := r.Session()

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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

	tx, err := spiffy.DB().Begin()

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
func (api API) getImagesAction(r *web.Ctx) web.Result {
	images, err := model.GetAllImages(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(images)
}

// POST "/api/images"
func (api API) createImageAction(r *web.Ctx) web.Result {
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
		return r.API().Result(existing)
	}

	session := r.Session()
	sessionUser := webutil.GetUser(session)
	image, err := CreateImageFromFile(session.UserID, !sessionUser.IsAdmin, postedFile.Contents, postedFile.FileName, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID))
	return r.API().Result(image)
}

// GET "/api/images/random/:count"
func (api API) getRandomImagesAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	images, err := model.GetRandomImages(count, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(images)
}

// GET "/api/images.search?query=<query>"
func (api API) searchImagesAction(r *web.Ctx) web.Result {
	contentRating := model.ContentRatingFilterDefault

	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && sessionUser.IsModerator {
		contentRating = model.ContentRatingFilterAll
	}

	query := r.Param("query")
	results, err := model.SearchImages(query, contentRating, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(results)
}

// GET "/api/images.search/random/:count?query=<query>"
func (api API) searchImagesRandomAction(r *web.Ctx) web.Result {
	contentRating := model.ContentRatingFilterDefault
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && sessionUser.IsModerator {
		contentRating = model.ContentRatingFilterAll
	}
	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	query := r.Param("query")
	results, err := model.SearchImagesWeightedRandom(query, contentRating, count, r.Tx())

	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(results)
}

// GET "/api/image/:image_id"
func (api API) getImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}
	return r.API().Result(image)
}

// PUT "/api/image/:image_id"
func (api API) updateImageAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	if !sessionUser.IsModerator {
		return r.API().NotAuthorized()
	}

	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	if image.IsZero() {
		return r.API().NotFound()
	}

	didUpdate := false

	updatedImage := model.Image{}
	err = r.PostBodyAsJSON(&updatedImage)

	if len(updatedImage.DisplayName) != 0 {
		image.DisplayName = updatedImage.DisplayName
		didUpdate = true
	}

	if updatedImage.ContentRating != image.ContentRating {
		image.ContentRating = updatedImage.ContentRating
		didUpdate = true
	}

	if didUpdate {
		r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(sessionUser.ID, model.ModerationVerbUpdate, model.ModerationObjectImage, image.UUID))
		err = spiffy.DB().Update(image)
		if err != nil {
			return r.API().InternalError(err)
		}
	}

	return r.API().Result(image)
}

// DELETE "/api/image/:image_id"
func (api API) deleteImageAction(r *web.Ctx) web.Result {
	session := r.Session()

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

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

	r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(session.UserID, model.ModerationVerbDelete, model.ModerationObjectImage, image.UUID))
	return r.API().OK()
}

// GET "/api/image.votes/:image_id"
func (api API) getLinksForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(voteSummaries)
}

// GET "/api/image.tags/:image_id"
func (api API) getTagsForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(results)
}

// GET "/api/tags"
func (api API) getRandomTagsAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	tags, err := model.GetRandomTags(count, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(tags)
}

// POST "/api/tags/random/:count"
func (api API) createTagsAction(r *web.Ctx) web.Result {
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

	session := r.Session()

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
		err = spiffy.DB().Create(tag)
		if err != nil {
			return r.API().InternalError(err)
		}
		r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(session.UserID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID))
		tags = append(tags, tag)
	}

	return r.API().Result(tags)
}

// GET "/api/tags"
func (api API) getTagsAction(r *web.Ctx) web.Result {
	tags, err := model.GetAllTags(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(tags)
}

// GET "/api/tags.search?query=<query>"
func (api API) searchTagsAction(r *web.Ctx) web.Result {
	query := r.Param("query")
	results, err := model.SearchTags(query, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(results)
}

// GET "/api/tags.search/random/:count?query=<query>"
func (api API) searchTagsRandomAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	query := r.Param("query")
	results, err := model.SearchTagsRandom(query, count, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().Result(results)
}

// GET "/api/tag/:tag_id"
func (api API) getTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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

	return r.API().Result(tag)
}

// DELETE "/api/tag/:tag_id"
func (api API) deleteTagAction(r *web.Ctx) web.Result {
	session := r.Session()

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
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

	err = model.DeleteTagAndVotesByID(tag.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(session.UserID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID))
	return r.API().OK()
}

// GET "/api/tag.images/:tag_id"
func (api API) getImagesForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(results)
}

// GET "/api/tag.votes/:tag_id"
func (api API) getLinksForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
	return r.API().Result(voteSummaries)
}

// DELETE "/api/link/:image_id/:tag_id"
func (api API) deleteLinkAction(r *web.Ctx) web.Result {
	session := r.Session()

	currentUser, err := model.GetUserByID(session.UserID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

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

	r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(session.UserID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID))

	return r.API().OK()
}

// GET "/api/teams"
func (api API) getTeamsAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	teams, err := model.GetAllSlackTeams(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(teams)
}

// GET "/api/team/:team_id"
func (api API) getTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	var team model.SlackTeam
	err = spiffy.DB().GetByIDInTx(&team, r.Tx(), teamID)
	if err != nil {
		return r.API().InternalError(err)
	}

	if team.IsZero() {
		return r.API().NotFound()
	}

	return r.API().Result(team)
}

// POST "/api/team"
func (api API) createTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	var team model.SlackTeam
	err := r.PostBodyAsJSON(&team)
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	err = spiffy.DB().CreateInTx(&team, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(team)
}

// PUT "/api/team/:team_id"
func (api API) updateTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	var team model.SlackTeam
	err = spiffy.DB().GetByIDInTx(&team, r.Tx(), teamID)
	if err != nil {
		return r.API().InternalError(err)
	}

	if team.IsZero() {
		return r.API().NotFound()
	}

	var updatedTeam model.SlackTeam
	err = r.PostBodyAsJSON(&updatedTeam)
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	updatedTeam.TeamID = teamID

	err = spiffy.DB().UpdateInTx(&updatedTeam, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(team)
}

// PATCH "/api/team/:team_id"
func (api API) patchTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	var team model.SlackTeam
	err = spiffy.DB().GetByIDInTx(&team, r.Tx(), teamID)
	if err != nil {
		return r.API().InternalError(err)
	}

	if team.IsZero() {
		return r.API().NotFound()
	}

	updates := map[string]interface{}{}
	err = r.PostBodyAsJSON(&updates)
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	err = util.Reflection.PatchObject(team, updates)
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	err = spiffy.DB().UpdateInTx(&team, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(team)
}

// DELETE "/api/team/:team_id"
func (api API) deleteTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	var team model.SlackTeam
	err = spiffy.DB().GetByIDInTx(&team, r.Tx(), teamID)
	if err != nil {
		return r.API().InternalError(err)
	}

	if team.IsZero() {
		return r.API().NotFound()
	}

	err = spiffy.DB().DeleteInTx(&team, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().OK()
}

// POST "/api/vote.up/:image_id/:tag_id"
func (api API) upvoteAction(r *web.Ctx) web.Result {
	return api.voteAction(true, r.Session(), r)
}

// POST "/api/vote.down/:image_id/:tag_id"
func (api API) downvoteAction(r *web.Ctx) web.Result {
	return api.voteAction(false, r.Session(), r)
}

func (api API) voteAction(isUpvote bool, session *web.Session, r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
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
		r.Logger().OnEvent(core.EventFlagModeration, model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID))
	}

	return r.API().OK()
}

// GET "/api/moderation.log/recent"
func (api API) getRecentModerationLogAction(r *web.Ctx) web.Result {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(moderationLog)
}

// GET "/api/moderation.log/pages/:count/:offset"
func (api API) getModerationLogByCountAndOffsetAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	log, err := model.GetModerationLogByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(log)
}

// GET "/api/search.history/recent"
func (api API) getRecentSearchHistoryAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	searchHistory, err := model.GetSearchHistory(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(searchHistory)
}

// GET "/api/search.history/pages/:count/:offset"
func (api API) getSearchHistoryByCountAndOffsetAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	count, err := r.RouteParamInt("count")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	searchHistory, err := model.GetSearchHistoryByCountAndOffset(count, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(searchHistory)
}

// GET "/api/stats"
func (api API) getSiteStatsAction(r *web.Ctx) web.Result {
	stats, err := viewmodel.GetSiteStats(r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(stats)
}

// GET "/api/image.stats/:image_id"
func (api API) getImageStatsAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	image, err := model.GetImageByUUID(imageUUID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}
	stats, err := viewmodel.GetImageStats(image.ID, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(stats)
}

// GET "/api/session.user"
func (api API) getCurrentUserAction(r *web.Ctx) web.Result {
	session := r.Session()

	cu := &viewmodel.CurrentUser{}
	if session == nil {
		cu.SetLoggedOut()
		return r.API().Result(cu)
	}
	cu.SetFromUser(webutil.GetUser(session))
	return r.API().Result(cu)
}

// GET "/api/session/:key"
func (api API) getSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	value, hasValue := session.State[key]
	if !hasValue {
		return r.API().NotFound()
	}
	return r.API().Result(value)
}

// POST "/api/session/:key"
// PUT "/api/session/:key"
func (api API) setSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	session.State[key], err = r.PostBodyAsString()
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().OK()
}

// DELETE "/api/session/:key"
func (api API) deleteSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}
	if _, hasKey := session.State[key]; !hasKey {
		return r.API().NotFound()
	}
	delete(session.State, key)
	return r.API().OK()
}

// GET "/api/jobs"
func (api API) getJobsStatusAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}
	status := chronometer.Default().Status()
	return r.API().Result(status)
}

// POST "/api/job/:job_id"
func (api API) runJobAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}
	jobID, err := r.RouteParam("job_id")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	err = chronometer.Default().RunJob(jobID)
	if err != nil {
		return r.API().InternalError(err)
	}
	return r.API().OK()
}

func (api API) getErrorsAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return r.API().NotAuthorized()
	}

	limit, err := r.RouteParamInt("limit")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return r.API().BadRequest(err.Error())
	}

	errors, err := model.GetAllErrorsWithLimitAndOffset(limit, offset, r.Tx())
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().Result(errors)
}

// POST "/api/logout"
func (api API) logoutAction(r *web.Ctx) web.Result {
	session := r.Session()

	if session == nil {
		return r.API().OK()
	}

	err := r.Auth().Logout(session, r)
	if err != nil {
		return r.API().InternalError(err)
	}

	return r.API().OK()
}
