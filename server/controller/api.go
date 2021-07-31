package controller

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/blend/go-sdk/cron"
	exception "github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/reflectutil"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

// APIs is the controller for api endpoints.
type APIs struct {
	Log    logger.Log
	Config *config.Giffy
	Model  *model.Manager
	OAuth  *oauth.Manager
	Files  *filemanager.FileManager
}

func (api APIs) awareMiddleware(extra ...web.Middleware) []web.Middleware {
	return append(extra, []web.Middleware{
		web.SessionAware,
		APIProviderAsDefault,
	}...)
}

func (api APIs) requiredMiddleware(extra ...web.Middleware) []web.Middleware {
	return append(extra, []web.Middleware{
		web.SessionRequired,
		APIProviderAsDefault,
	}...)
}

// Register adds the routes to the app.
func (api APIs) Register(app *web.App) {
	app.GET("/api/users", api.getUsersAction, api.requiredMiddleware()...)
	app.GET("/api/users.search", api.searchUsersAction, api.requiredMiddleware()...)
	app.GET("/api/users/pages/:count/:offset", api.getUsersByCountAndOffsetAction, api.requiredMiddleware()...)

	app.GET("/api/user/:user_id", api.getUserAction)
	app.PUT("/api/user/:user_id", api.updateUserAction, api.requiredMiddleware()...)
	app.GET("/api/user.images/:user_id", api.getUserImagesAction)
	app.GET("/api/user.moderation/:user_id", api.getModerationForUserAction)
	app.GET("/api/user.votes.image/:image_id", api.getVotesForUserForImageAction, api.requiredMiddleware()...)
	app.GET("/api/user.votes.tag/:tag_id", api.getVotesForUserForTagAction, api.requiredMiddleware()...)

	app.DELETE("/api/user.vote/:image_id/:tag_id", api.deleteUserVoteAction, api.requiredMiddleware()...)

	app.GET("/api/images", api.getImagesAction)
	app.POST("/api/images", api.createImageAction, api.requiredMiddleware()...)
	app.GET("/api/images/random/:count", api.getRandomImagesAction, api.awareMiddleware()...)
	app.GET("/api/images.search", api.searchImagesAction, api.awareMiddleware()...)
	app.GET("/api/images.search/random/:count", api.searchImagesRandomAction, api.awareMiddleware()...)

	app.GET("/api/image/:image_id", api.getImageAction)
	app.PUT("/api/image/:image_id", api.updateImageAction, api.requiredMiddleware()...)
	app.DELETE("/api/image/:image_id", api.deleteImageAction, api.requiredMiddleware()...)
	app.GET("/api/image.votes/:image_id", api.getLinksForImageAction)
	app.GET("/api/image.tags/:image_id", api.getTagsForImageAction)

	app.GET("/api/tags", api.getTagsAction)
	app.POST("/api/tags", api.createTagsAction, api.requiredMiddleware()...)
	app.GET("/api/tags/random/:count", api.getRandomTagsAction)
	app.GET("/api/tags.search", api.searchTagsAction)
	app.GET("/api/tags.search/random/:count", api.searchTagsRandomAction)

	app.GET("/api/tag/:tag_id", api.getTagAction)
	app.DELETE("/api/tag/:tag_id", api.deleteTagAction, api.requiredMiddleware()...)
	app.GET("/api/tag.images/:tag_id", api.getImagesForTagAction)
	app.GET("/api/tag.votes/:tag_id", api.getLinksForTagAction)

	app.GET("/api/teams", api.getTeamsAction, api.requiredMiddleware()...)
	app.GET("/api/team/:team_id", api.getTeamAction, api.requiredMiddleware()...)
	app.POST("/api/team/:team_id", api.createTeamAction, api.requiredMiddleware()...)
	app.PUT("/api/team/:team_id", api.updateTeamAction, api.requiredMiddleware()...)
	app.PATCH("/api/team/:team_id", api.patchTeamAction, api.requiredMiddleware()...)
	app.DELETE("/api/team/:team_id", api.deleteTeamAction, api.requiredMiddleware()...)

	app.DELETE("/api/link/:image_id/:tag_id", api.deleteLinkAction, api.requiredMiddleware()...)

	app.POST("/api/vote.up/:image_id/:tag_id", api.upvoteAction, api.requiredMiddleware()...)
	app.POST("/api/vote.down/:image_id/:tag_id", api.downvoteAction, api.requiredMiddleware()...)

	app.GET("/api/moderation.log/recent", api.getRecentModerationLogAction)
	app.GET("/api/moderation.log/pages/:count/:offset", api.getModerationLogByCountAndOffsetAction)

	app.GET("/api/search.history/recent", api.getRecentSearchHistoryAction, api.requiredMiddleware()...)
	app.GET("/api/search.history/pages/:count/:offset", api.getSearchHistoryByCountAndOffsetAction, api.requiredMiddleware()...)

	app.GET("/api/stats", api.getSiteStatsAction)
	app.GET("/api/image.stats/:image_id", api.getImageStatsAction)

	//session endpoints
	app.GET("/api/session.user", api.getCurrentUserAction, api.awareMiddleware()...)
	app.GET("/api/session/:key", api.getSessionKeyAction, api.requiredMiddleware()...)
	app.POST("/api/session/:key", api.setSessionKeyAction, api.requiredMiddleware()...)
	app.PUT("/api/session/:key", api.setSessionKeyAction, api.requiredMiddleware()...)
	app.DELETE("/api/session/:key", api.deleteSessionKeyAction, api.requiredMiddleware()...)

	//jobs
	app.GET("/api/jobs", api.getJobsStatusAction, api.requiredMiddleware()...)
	app.POST("/api/job/:job_id", api.runJobAction, api.requiredMiddleware()...)

	//errors
	app.GET("/api/errors/:limit/:offset", api.getErrorsAction, api.requiredMiddleware()...)

	// auth endpoints
	app.POST("/api/logout", api.logoutAction, api.requiredMiddleware()...)
}

// GET "/api/users"
func (api APIs) getUsersAction(r *web.Ctx) web.Result {
	user := GetUser(r.Session)
	if user != nil && !user.IsAdmin {
		return API(r).NotAuthorized()
	}

	users, err := api.Model.GetAllUsers(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(users)
}

// GET "/api/users.search"
func (api APIs) searchUsersAction(r *web.Ctx) web.Result {
	user := GetUser(r.Session)
	if user != nil && !user.IsAdmin {
		return API(r).NotAuthorized()
	}

	query := web.StringValue(r.Param("query"))
	users, err := api.Model.SearchUsers(r.Context(), query)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(users)
}

// GET "/api/users/pages/:count/:offset"
func (api APIs) getUsersByCountAndOffsetAction(r *web.Ctx) web.Result {
	user := GetUser(r.Session)
	if user != nil && !user.IsAdmin {
		return API(r).NotAuthorized()
	}

	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}
	offset, err := web.IntValue(r.RouteParam("offset"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	users, err := api.Model.GetUsersByCountAndOffset(r.Context(), count, offset)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(users)
}

// GET "/api/user/:user_id"
func (api APIs) getUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	user, err := api.Model.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if user.IsZero() {
		return API(r).NotFound()
	}

	return API(r).Result(user)
}

// PUT "/api/user/:user_id"
func (api APIs) updateUserAction(r *web.Ctx) web.Result {
	session := r.Session
	sessionUser := GetUser(session)
	if !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	user, err := api.Model.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if user.IsZero() {
		return API(r).NotFound()
	}

	var postedUser model.User
	err = r.PostBodyAsJSON(&postedUser)
	if err != nil {
		return API(r).BadRequest(exception.New("Post body was not properly formatted."))
	}

	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	postedUser.CreatedUTC = user.CreatedUTC
	postedUser.Username = user.Username

	if !user.IsAdmin && postedUser.IsAdmin {
		return API(r).BadRequest(exception.New("Cannot promote user to admin through the UI; this must be done in the db directly."))
	}

	if user.IsAdmin && !postedUser.IsAdmin {
		return API(r).BadRequest(exception.New("Cannot demote user from admin through the UI; this must be done in the db directly."))
	}

	if postedUser.IsAdmin && postedUser.IsBanned {
		return API(r).BadRequest(exception.New("Cannot ban admins."))
	}

	var moderationEntry *model.Moderation
	if !user.IsModerator && postedUser.IsModerator {
		moderationEntry = model.NewModeration(parseInt64(session.UserID), model.ModerationVerbPromoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsModerator && !postedUser.IsModerator {
		moderationEntry = model.NewModeration(parseInt64(session.UserID), model.ModerationVerbDemoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	}

	if !user.IsBanned && postedUser.IsBanned {
		moderationEntry = model.NewModeration(parseInt64(session.UserID), model.ModerationVerbBan, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsBanned && !postedUser.IsBanned {
		moderationEntry = model.NewModeration(parseInt64(session.UserID), model.ModerationVerbUnban, model.ModerationObjectUser, postedUser.UUID)
	}

	logger.MaybeTrigger(r.Context(), api.Log, moderationEntry)

	_, err = api.Model.Invoke(r.Context()).Update(&postedUser)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(postedUser)
}

// GET "/api/user.images/:user_id"
func (api APIs) getUserImagesAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	user, err := api.Model.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if user.IsZero() {
		return API(r).NotFound()
	}

	images, err := api.Model.GetImagesForUserID(r.Context(), user.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(viewmodel.WrapImages(images, api.Config))
}

// GET "/api/user.moderation/:user_id"
func (api APIs) getModerationForUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	user, err := api.Model.GetUserByUUID(r.Context(), userUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if user.IsZero() {
		return API(r).NotFound()
	}

	actions, err := api.Model.GetModerationForUserID(r.Context(), user.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(actions)
}

// GET "/api/user.votes.image/:image_id"
func (api APIs) getVotesForUserForImageAction(r *web.Ctx) web.Result {
	session := r.Session

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}
	votes, err := api.Model.GetVotesForUserForImage(r.Context(), parseInt64(session.UserID), image.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(votes)
}

// GET "/api/user.votes.tag/:tag_id"
func (api APIs) getVotesForUserForTagAction(r *web.Ctx) web.Result {
	session := r.Session

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}
	votes, err := api.Model.GetVotesForUserForTag(r.Context(), parseInt64(session.UserID), tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(votes)
}

// DELETE "/api/user.vote/:image_id/:tag_id"
func (api APIs) deleteUserVoteAction(r *web.Ctx) web.Result {
	session := r.Session

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	userID := parseInt64(session.UserID)

	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}

	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}

	vote, err := api.Model.GetVote(r.Context(), userID, image.ID, tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if vote.IsZero() {
		return API(r).NotFound()
	}

	// was it an upvote or downvote
	wasUpvote := vote.IsUpvote

	// adjust the vote summary ...
	voteSummary, err := api.Model.GetVoteSummary(r.Context(), image.ID, tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if wasUpvote {
		voteSummary.VotesFor--
	} else {
		voteSummary.VotesAgainst--
	}

	err = api.Model.SetVoteSummaryVoteCounts(r.Context(), image.ID, tag.ID, voteSummary.VotesFor, voteSummary.VotesAgainst)
	if err != nil {
		return API(r).InternalError(err)
	}

	err = api.Model.DeleteVote(r.Context(), userID, image.ID, tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).OK()
}

// GET "/api/images"
func (api APIs) getImagesAction(r *web.Ctx) web.Result {
	images, err := api.Model.GetAllImages(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(images)
}

// POST "/api/images"
func (api APIs) createImageAction(r *web.Ctx) web.Result {
	files, err := webutil.PostedFiles(r.Request)
	if err != nil {
		return API(r).BadRequest(fmt.Errorf("problem reading posted file: %v", err))
	}
	if len(files) == 0 {
		return API(r).BadRequest(fmt.Errorf("no files posted"))
	}

	postedFile := files[0]
	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := api.Model.GetImageByMD5(r.Context(), md5sum)
	if err != nil {
		return API(r).InternalError(err)
	}

	if !existing.IsZero() {
		return API(r).Result(existing)
	}

	session := r.Session
	sessionUser := GetUser(session)
	userID := parseInt64(session.UserID)
	image, err := CreateImageFromFile(r.Context(), api.Model, userID, !sessionUser.IsAdmin, postedFile.Contents, postedFile.FileName, api.Files)
	if err != nil {
		return API(r).InternalError(err)
	}

	logger.MaybeTrigger(r.Context(), api.Log, model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID))
	return API(r).Result(image)
}

// GET "/api/images/random/:count"
func (api APIs) getRandomImagesAction(r *web.Ctx) web.Result {
	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	images, err := api.Model.GetRandomImages(r.Context(), count)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(viewmodel.WrapImages(images, api.Config))
}

// GET "/api/images.search?query=<query>"
func (api APIs) searchImagesAction(r *web.Ctx) web.Result {
	contentRating := model.ContentRatingFilterDefault

	sessionUser := GetUser(r.Session)
	if sessionUser != nil && sessionUser.IsModerator {
		contentRating = model.ContentRatingFilterAll
	}

	query := web.StringValue(r.Param("query"))
	results, err := api.Model.SearchImages(r.Context(), query, contentRating)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(viewmodel.WrapImages(results, api.Config))
}

// GET "/api/images.search/random/:count?query=<query>"
func (api APIs) searchImagesRandomAction(r *web.Ctx) web.Result {
	contentRating := model.ContentRatingFilterDefault
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && sessionUser.IsModerator {
		contentRating = model.ContentRatingFilterAll
	}
	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	query := web.StringValue(r.Param("query"))
	results, err := api.Model.SearchImagesWeightedRandom(r.Context(), query, contentRating, count)

	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(viewmodel.WrapImages(results, api.Config))
}

// GET "/api/image/:image_id"
func (api APIs) getImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image == nil || image.IsZero() {
		return API(r).NotFound()
	}
	return API(r).Result(viewmodel.NewImage(*image, api.Config))
}

// PUT "/api/image/:image_id"
func (api APIs) updateImageAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	if !sessionUser.IsModerator {
		return API(r).NotAuthorized()
	}

	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}

	didUpdate := false

	updatedImage := model.Image{}
	err = r.PostBodyAsJSON(&updatedImage)
	if err != nil {
		return API(r).BadRequest(err)
	}

	if len(updatedImage.DisplayName) != 0 {
		image.DisplayName = updatedImage.DisplayName
		didUpdate = true
	}

	if updatedImage.ContentRating != image.ContentRating {
		image.ContentRating = updatedImage.ContentRating
		didUpdate = true
	}

	if didUpdate {
		logger.MaybeTrigger(r.Context(), api.Log, model.NewModeration(sessionUser.ID, model.ModerationVerbUpdate, model.ModerationObjectImage, image.UUID))
		_, err = api.Model.Invoke(r.Context()).Update(image)
		if err != nil {
			return API(r).InternalError(err)
		}
	}

	return API(r).Result(image)
}

// DELETE "/api/image/:image_id"
func (api APIs) deleteImageAction(r *web.Ctx) web.Result {
	session := r.Session

	currentUser, err := api.Model.GetUserByID(r.Context(), parseInt64(session.UserID))
	if err != nil {
		return API(r).InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}
	if !currentUser.IsModerator && image.CreatedBy != currentUser.ID {
		return API(r).NotAuthorized()
	}

	//delete from s3 (!!)
	err = api.Files.DeleteFile(api.Files.NewLocationFromKey(image.S3Key))
	if err != nil {
		return API(r).InternalError(err)
	}

	err = api.Model.DeleteImageByID(r.Context(), image.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	logger.MaybeTrigger(
		r.Context(),
		api.Log,
		model.NewModeration(parseInt64(session.UserID), model.ModerationVerbDelete, model.ModerationObjectImage, image.UUID),
	)
	return API(r).OK()
}

// GET "/api/image.votes/:image_id"
func (api APIs) getLinksForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}
	voteSummaries, err := api.Model.GetVoteSummariesForImage(r.Context(), image.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(voteSummaries)
}

// GET "/api/image.tags/:image_id"
func (api APIs) getTagsForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}

	results, err := api.Model.GetTagsForImageID(r.Context(), image.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(results)
}

// GET "/api/tags"
func (api APIs) getRandomTagsAction(r *web.Ctx) web.Result {
	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	tags, err := api.Model.GetRandomTags(r.Context(), count)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(tags)
}

// POST "/api/tags/random/:count"
func (api APIs) createTagsAction(r *web.Ctx) web.Result {
	args := viewmodel.CreateTagArgs{}
	err := r.PostBodyAsJSON(&args)
	if err != nil {
		return API(r).BadRequest(err)
	}

	if len(args) == 0 {
		return API(r).BadRequest(fmt.Errorf("empty post body, please submit an array of strings"))
	}

	var tagValues []string
	for _, value := range args {
		tagValue := model.CleanTagValue(value)
		if len(tagValue) == 0 {
			return API(r).BadRequest(fmt.Errorf("`tag_value` be in the form [a-z,A-Z,0-9]+"))
		}

		tagValues = append(tagValues, tagValue)
	}

	session := r.Session

	tags := []*model.Tag{}
	for _, tagValue := range tagValues {
		existingTag, err := api.Model.GetTagByValue(r.Context(), tagValue)

		if err != nil {
			return API(r).InternalError(err)
		}

		if !existingTag.IsZero() {
			tags = append(tags, existingTag)
			continue
		}

		userID := parseInt64(session.UserID)
		tag := model.NewTag(userID, tagValue)
		err = api.Model.Invoke(r.Context()).Create(tag)
		if err != nil {
			return API(r).InternalError(err)
		}
		logger.MaybeTrigger(r.Context(), api.Log, model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID))
		tags = append(tags, tag)
	}

	return API(r).Result(tags)
}

// GET "/api/tags"
func (api APIs) getTagsAction(r *web.Ctx) web.Result {
	tags, err := api.Model.GetAllTags(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(tags)
}

// GET "/api/tags.search?query=<query>"
func (api APIs) searchTagsAction(r *web.Ctx) web.Result {
	query := web.StringValue(r.Param("query"))
	results, err := api.Model.SearchTags(r.Context(), query)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(results)
}

// GET "/api/tags.search/random/:count?query=<query>"
func (api APIs) searchTagsRandomAction(r *web.Ctx) web.Result {
	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}
	query := web.StringValue(r.Param("query"))
	results, err := api.Model.SearchTagsRandom(r.Context(), query, count)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(results)
}

// GET "/api/tag/:tag_id"
func (api APIs) getTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if tag.IsZero() {
		tag, err = api.Model.GetTagByValue(r.Context(), tagUUID)
		if err != nil {
			return API(r).InternalError(err)
		}
		if tag.IsZero() {
			return API(r).NotFound()
		}
	}

	return API(r).Result(tag)
}

// DELETE "/api/tag/:tag_id"
func (api APIs) deleteTagAction(r *web.Ctx) web.Result {
	session := r.Session
	userID := parseInt64(session.UserID)

	currentUser, err := api.Model.GetUserByID(r.Context(), userID)
	if err != nil {
		return API(r).InternalError(err)
	}

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return API(r).NotAuthorized()
	}

	err = api.Model.DeleteTagAndVotesByID(r.Context(), tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	logger.MaybeTrigger(r.Context(), api.Log, model.NewModeration(userID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID))
	return API(r).OK()
}

// GET "/api/tag.images/:tag_id"
func (api APIs) getImagesForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}

	results, err := api.Model.GetImagesForTagID(r.Context(), tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(viewmodel.WrapImages(results, api.Config))
}

// GET "/api/tag.votes/:tag_id"
func (api APIs) getLinksForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}
	voteSummaries, err := api.Model.GetVoteSummariesForTag(r.Context(), tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).Result(voteSummaries)
}

// DELETE "/api/link/:image_id/:tag_id"
func (api APIs) deleteLinkAction(r *web.Ctx) web.Result {
	session := r.Session
	userID := parseInt64(session.UserID)

	currentUser, err := api.Model.GetUserByID(r.Context(), userID)
	if err != nil {
		return API(r).InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}

	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return API(r).NotAuthorized()
	}

	err = api.Model.DeleteVoteSummary(r.Context(), image.ID, tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	logger.MaybeTrigger(r.Context(), api.Log, model.NewModeration(userID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID))
	return API(r).OK()
}

// GET "/api/teams"
func (api APIs) getTeamsAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	teams, err := api.Model.GetAllSlackTeams(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(teams)
}

// GET "/api/team/:team_id"
func (api APIs) getTeamAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	var team model.SlackTeam
	_, err = api.Model.Invoke(r.Context()).Get(&team, teamID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if team.IsZero() {
		return API(r).NotFound()
	}

	return API(r).Result(team)
}

// POST "/api/team"
func (api APIs) createTeamAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	var team model.SlackTeam
	err := r.PostBodyAsJSON(&team)
	if err != nil {
		return API(r).BadRequest(err)
	}

	err = api.Model.Invoke(r.Context()).Create(&team)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(team)
}

// PUT "/api/team/:team_id"
func (api APIs) updateTeamAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	var team model.SlackTeam
	_, err = api.Model.Invoke(r.Context()).Get(&team, teamID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if team.IsZero() {
		return API(r).NotFound()
	}

	var updatedTeam model.SlackTeam
	err = r.PostBodyAsJSON(&updatedTeam)
	if err != nil {
		return API(r).BadRequest(err)
	}

	updatedTeam.TeamID = teamID

	_, err = api.Model.Invoke(r.Context()).Update(&updatedTeam)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(team)
}

// PATCH "/api/team/:team_id"
func (api APIs) patchTeamAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	var team model.SlackTeam
	_, err = api.Model.Invoke(r.Context()).Get(&team, teamID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if team.IsZero() {
		return API(r).NotFound()
	}

	updates := map[string]interface{}{}
	err = r.PostBodyAsJSON(&updates)
	if err != nil {
		return API(r).BadRequest(err)
	}

	err = reflectutil.Patch(team, updates)
	if err != nil {
		return API(r).BadRequest(err)
	}

	_, err = api.Model.Invoke(r.Context()).Update(&team)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(team)
}

// DELETE "/api/team/:team_id"
func (api APIs) deleteTeamAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	var team model.SlackTeam
	_, err = api.Model.Invoke(r.Context()).Get(&team, teamID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if team.IsZero() {
		return API(r).NotFound()
	}

	_, err = api.Model.Invoke(r.Context()).Delete(&team)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).OK()
}

// POST "/api/vote.up/:image_id/:tag_id"
func (api APIs) upvoteAction(r *web.Ctx) web.Result {
	return api.voteAction(true, r.Session, r)
}

// POST "/api/vote.down/:image_id/:tag_id"
func (api APIs) downvoteAction(r *web.Ctx) web.Result {
	return api.voteAction(false, r.Session, r)
}

func (api APIs) voteAction(isUpvote bool, session *web.Session, r *web.Ctx) web.Result {
	userID := parseInt64(session.UserID)

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	tag, err := api.Model.GetTagByUUID(r.Context(), tagUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if tag.IsZero() {
		return API(r).NotFound()
	}

	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	if image.IsZero() {
		return API(r).NotFound()
	}

	existingUserVote, err := api.Model.GetVote(r.Context(), userID, image.ID, tag.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return API(r).OK()
	}

	didCreate, err := api.Model.CreateOrUpdateVote(r.Context(), userID, image.ID, tag.ID, isUpvote)
	if err != nil {
		return API(r).InternalError(err)
	}

	if didCreate {
		logger.MaybeTrigger(
			r.Context(),
			api.Log,
			model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID),
		)
	}

	return API(r).OK()
}

// GET "/api/moderation.log/recent"
func (api APIs) getRecentModerationLogAction(r *web.Ctx) web.Result {
	moderationLog, err := api.Model.GetModerationsByTime(r.Context(), time.Now().UTC().AddDate(0, 0, -1))
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(moderationLog)
}

// GET "/api/moderation.log/pages/:count/:offset"
func (api APIs) getModerationLogByCountAndOffsetAction(r *web.Ctx) web.Result {
	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}
	offset, err := web.IntValue(r.RouteParam("offset"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	log, err := api.Model.GetModerationLogByCountAndOffset(r.Context(), count, offset)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(log)
}

// GET "/api/search.history/recent"
func (api APIs) getRecentSearchHistoryAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	searchHistory, err := api.Model.GetSearchHistory(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(searchHistory)
}

// GET "/api/search.history/pages/:count/:offset"
func (api APIs) getSearchHistoryByCountAndOffsetAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	count, err := web.IntValue(r.RouteParam("count"))
	if err != nil {
		return API(r).BadRequest(err)
	}
	offset, err := web.IntValue(r.RouteParam("offset"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	searchHistory, err := api.Model.GetSearchHistoryByCountAndOffset(r.Context(), count, offset)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(searchHistory)
}

// GET "/api/stats"
func (api APIs) getSiteStatsAction(r *web.Ctx) web.Result {
	stats, err := api.Model.GetSiteStats(r.Context())
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(stats)
}

// GET "/api/image.stats/:image_id"
func (api APIs) getImageStatsAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return API(r).BadRequest(err)
	}
	image, err := api.Model.GetImageByUUID(r.Context(), imageUUID)
	if err != nil {
		return API(r).InternalError(err)
	}
	stats, err := api.Model.GetImageStats(r.Context(), image.ID)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(stats)
}

// GET "/api/session.user"
func (api APIs) getCurrentUserAction(r *web.Ctx) web.Result {
	session := r.Session

	url, err := api.OAuth.OAuthURL(r.Request)
	if err != nil {
		return API(r).InternalError(err)
	}
	cu := &viewmodel.CurrentUser{
		GoogleLoginURL: url,
		SlackLoginURL:  external.SlackAuthURL(api.Config),
	}
	if session == nil {
		return API(r).Result(cu)
	}

	cu.IsLoggedIn = true
	cu.SetFromUser(GetUser(session))
	return API(r).Result(cu)
}

// GET "/api/session/:key"
func (api APIs) getSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session

	key, err := r.RouteParam("key")
	if err != nil {
		return API(r).BadRequest(err)
	}
	value, hasValue := session.State[key]
	if !hasValue {
		return API(r).NotFound()
	}
	return API(r).Result(value)
}

// POST "/api/session/:key"
// PUT "/api/session/:key"
func (api APIs) setSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session

	key, err := r.RouteParam("key")
	if err != nil {
		return API(r).BadRequest(err)
	}
	session.State[key], err = r.PostBodyAsString()
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).OK()
}

// DELETE "/api/session/:key"
func (api APIs) deleteSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session

	key, err := r.RouteParam("key")
	if err != nil {
		return API(r).BadRequest(err)
	}
	if _, hasKey := session.State[key]; !hasKey {
		return API(r).NotFound()
	}
	delete(session.State, key)
	return API(r).OK()
}

// GET "/api/jobs"
func (api APIs) getJobsStatusAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}
	status := cron.Default().State()
	return API(r).Result(status)
}

// POST "/api/job/:job_id"
func (api APIs) runJobAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}
	jobID, err := r.RouteParam("job_id")
	if err != nil {
		return API(r).BadRequest(err)
	}

	_, _, err = cron.Default().RunJob(jobID)
	if err != nil {
		return API(r).InternalError(err)
	}
	return API(r).OK()
}

func (api APIs) getErrorsAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if sessionUser != nil && !sessionUser.IsAdmin {
		return API(r).NotAuthorized()
	}

	limit, err := web.IntValue(r.RouteParam("limit"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	offset, err := web.IntValue(r.RouteParam("offset"))
	if err != nil {
		return API(r).BadRequest(err)
	}

	errors, err := api.Model.GetAllErrorsWithLimitAndOffset(r.Context(), limit, offset)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).Result(errors)
}

// POST "/api/logout"
func (api APIs) logoutAction(r *web.Ctx) web.Result {
	session := r.Session

	if session == nil {
		return API(r).OK()
	}

	err := r.Auth.Logout(r)
	if err != nil {
		return API(r).InternalError(err)
	}

	return API(r).OK()
}
