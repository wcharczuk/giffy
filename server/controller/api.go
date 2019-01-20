package controller

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
	"github.com/wcharczuk/giffy/server/webutil"
)

// API is the controller for api endpoints.
type API struct {
	Config *config.Giffy
	Model *model.Manager
	OAuth  *oauth.Manager
	Files  *filemanager.FileManager
}

// Register adds the routes to the app.
func (api API) Register(app *web.App) {
	app.GET("/api/users", api.getUsersAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/users.search", api.searchUsersAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/users/pages/:count/:offset", api.getUsersByCountAndOffsetAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.GET("/api/user/:user_id", api.getUserAction)
	app.PUT("/api/user/:user_id", api.updateUserAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/user.images/:user_id", api.getUserImagesAction)
	app.GET("/api/user.moderation/:user_id", api.getModerationForUserAction)
	app.GET("/api/user.votes.image/:image_id", api.getVotesForUserForImageAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/user.votes.tag/:tag_id", api.getVotesForUserForTagAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.DELETE("/api/user.vote/:image_id/:tag_id", api.deleteUserVoteAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.GET("/api/images", api.getImagesAction)
	app.POST("/api/images", api.createImageAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/images/random/:count", api.getRandomImagesAction, web.SessionAware, webutil.APIProviderAsDefault)
	app.GET("/api/images.search", api.searchImagesAction, web.SessionAware, webutil.APIProviderAsDefault)
	app.GET("/api/images.search/random/:count", api.searchImagesRandomAction, web.SessionAware, webutil.APIProviderAsDefault)

	app.GET("/api/image/:image_id", api.getImageAction)
	app.PUT("/api/image/:image_id", api.updateImageAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.DELETE("/api/image/:image_id", api.deleteImageAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/image.votes/:image_id", api.getLinksForImageAction)
	app.GET("/api/image.tags/:image_id", api.getTagsForImageAction)

	app.GET("/api/tags", api.getTagsAction)
	app.POST("/api/tags", api.createTagsAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/tags/random/:count", api.getRandomTagsAction)
	app.GET("/api/tags.search", api.searchTagsAction)
	app.GET("/api/tags.search/random/:count", api.searchTagsRandomAction)

	app.GET("/api/tag/:tag_id", api.getTagAction)
	app.DELETE("/api/tag/:tag_id", api.deleteTagAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/tag.images/:tag_id", api.getImagesForTagAction)
	app.GET("/api/tag.votes/:tag_id", api.getLinksForTagAction)

	app.GET("/api/teams", api.getTeamsAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/team/:team_id", api.getTeamAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.POST("/api/team/:team_id", api.createTeamAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.PUT("/api/team/:team_id", api.updateTeamAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.PATCH("/api/team/:team_id", api.patchTeamAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.DELETE("/api/team/:team_id", api.deleteTeamAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.DELETE("/api/link/:image_id/:tag_id", api.deleteLinkAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.POST("/api/vote.up/:image_id/:tag_id", api.upvoteAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.POST("/api/vote.down/:image_id/:tag_id", api.downvoteAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.GET("/api/moderation.log/recent", api.getRecentModerationLogAction)
	app.GET("/api/moderation.log/pages/:count/:offset", api.getModerationLogByCountAndOffsetAction)

	app.GET("/api/search.history/recent", api.getRecentSearchHistoryAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.GET("/api/search.history/pages/:count/:offset", api.getSearchHistoryByCountAndOffsetAction, web.SessionRequired, webutil.APIProviderAsDefault)

	app.GET("/api/stats", api.getSiteStatsAction)
	app.GET("/api/image.stats/:image_id", api.getImageStatsAction)

	//session endpoints
	app.GET("/api/session.user", api.getCurrentUserAction, web.SessionAware, webutil.APIProviderAsDefault)
	app.GET("/api/session/:key", api.getSessionKeyAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.POST("/api/session/:key", api.setSessionKeyAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.PUT("/api/session/:key", api.setSessionKeyAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.DELETE("/api/session/:key", api.deleteSessionKeyAction, web.SessionRequired, webutil.APIProviderAsDefault)

	//jobs
	app.GET("/api/jobs", api.getJobsStatusAction, web.SessionRequired, webutil.APIProviderAsDefault)
	app.POST("/api/job/:job_id", api.runJobAction, web.SessionRequired, webutil.APIProviderAsDefault)

	//errors
	app.GET("/api/errors/:limit/:offset", api.getErrorsAction, web.SessionRequired, webutil.APIProviderAsDefault)

	// auth endpoints
	app.POST("/api/logout", api.logoutAction, web.SessionRequired, webutil.APIProviderAsDefault)
}

// GET "/api/users"
func (api API) getUsersAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	users, err := api.Model.GetAllUsers(r.Context())
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(users)
}

// GET "/api/users.search"
func (api API) searchUsersAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	query := r.ParamString("query")
	users, err := model.SearchUsers(query, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(users)
}

// GET "/api/users/pages/:count/:offset"
func (api API) getUsersByCountAndOffsetAction(r *web.Ctx) web.Result {
	user := webutil.GetUser(r.Session())
	if user != nil && !user.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	users, err := model.GetUsersByCountAndOffset(count, offset, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(users)
}

// GET "/api/user/:user_id"
func (api API) getUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	user, err := model.GetUserByUUID(userUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if user.IsZero() {
		return webutil.API(r).NotFound()
	}

	return webutil.API(r).Result(user)
}

// PUT "/api/user/:user_id"
func (api API) updateUserAction(r *web.Ctx) web.Result {
	session := r.Session()
	sessionUser := webutil.GetUser(session)
	if !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	user, err := model.GetUserByUUID(userUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if user.IsZero() {
		return webutil.API(r).NotFound()
	}

	var postedUser model.User
	err = r.PostBodyAsJSON(&postedUser)
	if err != nil {
		return webutil.API(r).BadRequest(exception.New("Post body was not properly formatted."))
	}

	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	postedUser.CreatedUTC = user.CreatedUTC
	postedUser.Username = user.Username

	if !user.IsAdmin && postedUser.IsAdmin {
		return webutil.API(r).BadRequest(exception.New("Cannot promote user to admin through the UI; this must be done in the db directly."))
	}

	if user.IsAdmin && !postedUser.IsAdmin {
		return webutil.API(r).BadRequest(exception.New("Cannot demote user from admin through the UI; this must be done in the db directly."))
	}

	if postedUser.IsAdmin && postedUser.IsBanned {
		return webutil.API(r).BadRequest(exception.New("Cannot ban admins."))
	}

	var moderationEntry *model.Moderation
	if !user.IsModerator && postedUser.IsModerator {
		moderationEntry = model.NewModeration(util.Parse.Int64(session.UserID), model.ModerationVerbPromoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsModerator && !postedUser.IsModerator {
		moderationEntry = model.NewModeration(util.Parse.Int64(session.UserID), model.ModerationVerbDemoteAsModerator, model.ModerationObjectUser, postedUser.UUID)
	}

	if !user.IsBanned && postedUser.IsBanned {
		moderationEntry = model.NewModeration(util.Parse.Int64(session.UserID), model.ModerationVerbBan, model.ModerationObjectUser, postedUser.UUID)
	} else if user.IsBanned && !postedUser.IsBanned {
		moderationEntry = model.NewModeration(util.Parse.Int64(session.UserID), model.ModerationVerbUnban, model.ModerationObjectUser, postedUser.UUID)
	}

	r.Logger().Trigger(moderationEntry)

	err = model.DB().Update(&postedUser)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(postedUser)
}

// GET "/api/user.images/:user_id"
func (api API) getUserImagesAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	user, err := model.GetUserByUUID(userUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if user.IsZero() {
		return webutil.API(r).NotFound()
	}

	images, err := model.GetImagesForUserID(user.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(viewmodel.WrapImages(images, api.Config))
}

// GET "/api/user.moderation/:user_id"
func (api API) getModerationForUserAction(r *web.Ctx) web.Result {
	userUUID, err := r.RouteParam("user_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	user, err := model.GetUserByUUID(userUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if user.IsZero() {
		return webutil.API(r).NotFound()
	}

	actions, err := model.GetModerationForUserID(user.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(actions)
}

// GET "/api/user.votes.image/:image_id"
func (api API) getVotesForUserForImageAction(r *web.Ctx) web.Result {
	session := r.Session()

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}
	votes, err := model.GetVotesForUserForImage(util.Parse.Int64(session.UserID), image.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(votes)
}

// GET "/api/user.votes.tag/:tag_id"
func (api API) getVotesForUserForTagAction(r *web.Ctx) web.Result {
	session := r.Session()

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}
	votes, err := model.GetVotesForUserForTag(util.Parse.Int64(session.UserID), tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(votes)
}

// DELETE "/api/user.vote/:image_id/:tag_id"
func (api API) deleteUserVoteAction(r *web.Ctx) web.Result {
	session := r.Session()

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	userID := util.Parse.Int64(session.UserID)

	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}

	tx, err := model.DB().Begin()

	vote, err := model.GetVote(userID, image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return webutil.API(r).InternalError(err)
	}
	if vote.IsZero() {
		tx.Rollback()
		return webutil.API(r).NotFound()
	}

	// was it an upvote or downvote
	wasUpvote := vote.IsUpvote

	// adjust the vote summary ...
	voteSummary, err := model.GetVoteSummary(image.ID, tag.ID, tx)
	if err != nil {
		tx.Rollback()
		return webutil.API(r).InternalError(err)
	}

	if wasUpvote {
		voteSummary.VotesFor--
	} else {
		voteSummary.VotesAgainst--
	}

	err = model.SetVoteSummaryVoteCounts(image.ID, tag.ID, voteSummary.VotesFor, voteSummary.VotesAgainst, tx)
	if err != nil {
		tx.Rollback()
		return webutil.API(r).InternalError(err)
	}

	err = model.DeleteVote(userID, image.ID, tag.ID, web.Tx(r))
	if err != nil {
		tx.Rollback()
		return webutil.API(r).InternalError(err)
	}

	tx.Commit()
	return webutil.API(r).OK()
}

// GET "/api/images"
func (api API) getImagesAction(r *web.Ctx) web.Result {
	images, err := model.GetAllImages(web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(images)
}

// POST "/api/images"
func (api API) createImageAction(r *web.Ctx) web.Result {
	files, filesErr := r.PostedFiles()
	if filesErr != nil {
		return webutil.API(r).BadRequest(fmt.Errorf("problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return webutil.API(r).BadRequest(fmt.Errorf("no files posted"))
	}

	postedFile := files[0]
	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if !existing.IsZero() {
		return webutil.API(r).Result(existing)
	}

	session := r.Session()
	sessionUser := webutil.GetUser(session)
	userID := util.Parse.Int64(session.UserID)
	image, err := CreateImageFromFile(userID, !sessionUser.IsAdmin, postedFile.Contents, postedFile.FileName, api.Files, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	r.Logger().Trigger(model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID))
	return webutil.API(r).Result(image)
}

// GET "/api/images/random/:count"
func (api API) getRandomImagesAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	images, err := model.GetRandomImages(count, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(viewmodel.WrapImages(images, api.Config))
}

// GET "/api/images.search?query=<query>"
func (api API) searchImagesAction(r *web.Ctx) web.Result {
	contentRating := model.ContentRatingFilterDefault

	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && sessionUser.IsModerator {
		contentRating = model.ContentRatingFilterAll
	}

	query := r.ParamString("query")
	results, err := model.SearchImages(query, contentRating, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(viewmodel.WrapImages(results, api.Config))
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
		return webutil.API(r).BadRequest(err)
	}

	query := r.ParamString("query")
	results, err := model.SearchImagesWeightedRandom(query, contentRating, count, web.Tx(r))

	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(viewmodel.WrapImages(results, api.Config))
}

// GET "/api/image/:image_id"
func (api API) getImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image == nil || image.IsZero() {
		return webutil.API(r).NotFound()
	}
	return webutil.API(r).Result(viewmodel.NewImage(*image, api.Config))
}

// PUT "/api/image/:image_id"
func (api API) updateImageAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	if !sessionUser.IsModerator {
		return webutil.API(r).NotAuthorized()
	}

	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
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
		r.Logger().Trigger(model.NewModeration(sessionUser.ID, model.ModerationVerbUpdate, model.ModerationObjectImage, image.UUID))
		err = model.DB().Update(image)
		if err != nil {
			return webutil.API(r).InternalError(err)
		}
	}

	return webutil.API(r).Result(image)
}

// DELETE "/api/image/:image_id"
func (api API) deleteImageAction(r *web.Ctx) web.Result {
	session := r.Session()

	currentUser, err := model.GetUserByID(util.Parse.Int64(session.UserID), web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}
	if !currentUser.IsModerator && image.CreatedBy != currentUser.ID {
		return webutil.API(r).NotAuthorized()
	}

	//delete from s3 (!!)
	err = api.Files.DeleteFile(api.Files.NewLocationFromKey(image.S3Key))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	err = model.DeleteImageByID(image.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	r.Logger().Trigger(model.NewModeration(util.Parse.Int64(session.UserID), model.ModerationVerbDelete, model.ModerationObjectImage, image.UUID))
	return webutil.API(r).OK()
}

// GET "/api/image.votes/:image_id"
func (api API) getLinksForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForImage(image.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(voteSummaries)
}

// GET "/api/image.tags/:image_id"
func (api API) getTagsForImageAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}

	results, err := model.GetTagsForImageID(image.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(results)
}

// GET "/api/tags"
func (api API) getRandomTagsAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	tags, err := model.GetRandomTags(count, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(tags)
}

// POST "/api/tags/random/:count"
func (api API) createTagsAction(r *web.Ctx) web.Result {
	args := viewmodel.CreateTagArgs{}
	err := r.PostBodyAsJSON(&args)
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	if len(args) == 0 {
		return webutil.API(r).BadRequest(fmt.Errorf("empty post body, please submit an array of strings"))
	}

	var tagValues []string
	for _, value := range args {
		tagValue := model.CleanTagValue(value)
		if len(tagValue) == 0 {
			return webutil.API(r).BadRequest(fmt.Errorf("`tag_value` be in the form [a-z,A-Z,0-9]+"))
		}

		tagValues = append(tagValues, tagValue)
	}

	session := r.Session()

	tags := []*model.Tag{}
	for _, tagValue := range tagValues {
		existingTag, err := model.GetTagByValue(tagValue, web.Tx(r))

		if err != nil {
			return webutil.API(r).InternalError(err)
		}

		if !existingTag.IsZero() {
			tags = append(tags, existingTag)
			continue
		}

		userID := util.Parse.Int64(session.UserID)
		tag := model.NewTag(userID, tagValue)
		err = model.DB().Create(tag)
		if err != nil {
			return webutil.API(r).InternalError(err)
		}
		r.Logger().Trigger(model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectTag, tag.UUID))
		tags = append(tags, tag)
	}

	return webutil.API(r).Result(tags)
}

// GET "/api/tags"
func (api API) getTagsAction(r *web.Ctx) web.Result {
	tags, err := model.GetAllTags(web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(tags)
}

// GET "/api/tags.search?query=<query>"
func (api API) searchTagsAction(r *web.Ctx) web.Result {
	query := r.ParamString("query")
	results, err := model.SearchTags(query, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(results)
}

// GET "/api/tags.search/random/:count?query=<query>"
func (api API) searchTagsRandomAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	query := r.ParamString("query")
	results, err := model.SearchTagsRandom(query, count, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(results)
}

// GET "/api/tag/:tag_id"
func (api API) getTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if tag.IsZero() {
		tag, err = model.GetTagByValue(tagUUID, web.Tx(r))
		if err != nil {
			return webutil.API(r).InternalError(err)
		}
		if tag.IsZero() {
			return webutil.API(r).NotFound()
		}
	}

	return webutil.API(r).Result(tag)
}

// DELETE "/api/tag/:tag_id"
func (api API) deleteTagAction(r *web.Ctx) web.Result {
	session := r.Session()
	userID := util.Parse.Int64(session.UserID)

	currentUser, err := model.GetUserByID(userID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return webutil.API(r).NotAuthorized()
	}

	err = model.DeleteTagAndVotesByID(tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	r.Logger().Trigger(model.NewModeration(userID, model.ModerationVerbDelete, model.ModerationObjectTag, tag.UUID))
	return webutil.API(r).OK()
}

// GET "/api/tag.images/:tag_id"
func (api API) getImagesForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}

	results, err := model.GetImagesForTagID(tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(viewmodel.WrapImages(results, api.Config))
}

// GET "/api/tag.votes/:tag_id"
func (api API) getLinksForTagAction(r *web.Ctx) web.Result {
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}
	voteSummaries, err := model.GetVoteSummariesForTag(tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).Result(voteSummaries)
}

// DELETE "/api/link/:image_id/:tag_id"
func (api API) deleteLinkAction(r *web.Ctx) web.Result {
	session := r.Session()
	userID := util.Parse.Int64(session.UserID)

	currentUser, err := model.GetUserByID(userID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}

	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}
	if !currentUser.IsModerator && tag.CreatedBy != currentUser.ID {
		return webutil.API(r).NotAuthorized()
	}

	err = model.DeleteVoteSummary(image.ID, tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	r.Logger().Trigger(model.NewModeration(userID, model.ModerationVerbDelete, model.ModerationObjectLink, imageUUID, tagUUID))

	return webutil.API(r).OK()
}

// GET "/api/teams"
func (api API) getTeamsAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	teams, err := model.GetAllSlackTeams(web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(teams)
}

// GET "/api/team/:team_id"
func (api API) getTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	var team model.SlackTeam
	err = model.DB().GetInTx(&team, web.Tx(r), teamID)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if team.IsZero() {
		return webutil.API(r).NotFound()
	}

	return webutil.API(r).Result(team)
}

// POST "/api/team"
func (api API) createTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	var team model.SlackTeam
	err := r.PostBodyAsJSON(&team)
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	err = model.DB().CreateInTx(&team, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(team)
}

// PUT "/api/team/:team_id"
func (api API) updateTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	var team model.SlackTeam
	err = model.DB().GetInTx(&team, web.Tx(r), teamID)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if team.IsZero() {
		return webutil.API(r).NotFound()
	}

	var updatedTeam model.SlackTeam
	err = r.PostBodyAsJSON(&updatedTeam)
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	updatedTeam.TeamID = teamID

	err = model.DB().UpdateInTx(&updatedTeam, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(team)
}

// PATCH "/api/team/:team_id"
func (api API) patchTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	var team model.SlackTeam
	err = model.DB().GetInTx(&team, web.Tx(r), teamID)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if team.IsZero() {
		return webutil.API(r).NotFound()
	}

	updates := map[string]interface{}{}
	err = r.PostBodyAsJSON(&updates)
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	err = util.Reflection.PatchObject(team, updates)
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	err = model.DB().UpdateInTx(&team, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(team)
}

// DELETE "/api/team/:team_id"
func (api API) deleteTeamAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	teamID, err := r.RouteParam("team_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	var team model.SlackTeam
	err = model.DB().GetInTx(&team, web.Tx(r), teamID)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if team.IsZero() {
		return webutil.API(r).NotFound()
	}

	err = model.DB().DeleteInTx(&team, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).OK()
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
	userID := util.Parse.Int64(session.UserID)

	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	tagUUID, err := r.RouteParam("tag_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	tag, err := model.GetTagByUUID(tagUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if tag.IsZero() {
		return webutil.API(r).NotFound()
	}

	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	if image.IsZero() {
		return webutil.API(r).NotFound()
	}

	existingUserVote, err := model.GetVote(userID, image.ID, tag.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if !existingUserVote.IsZero() {
		return webutil.API(r).OK()
	}

	didCreate, err := model.CreateOrUpdateVote(userID, image.ID, tag.ID, isUpvote, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	if didCreate {
		r.Logger().Trigger(model.NewModeration(userID, model.ModerationVerbCreate, model.ModerationObjectLink, imageUUID, tagUUID))
	}

	return webutil.API(r).OK()
}

// GET "/api/moderation.log/recent"
func (api API) getRecentModerationLogAction(r *web.Ctx) web.Result {
	moderationLog, err := model.GetModerationsByTime(time.Now().UTC().AddDate(0, 0, -1), web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(moderationLog)
}

// GET "/api/moderation.log/pages/:count/:offset"
func (api API) getModerationLogByCountAndOffsetAction(r *web.Ctx) web.Result {
	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	log, err := model.GetModerationLogByCountAndOffset(count, offset, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(log)
}

// GET "/api/search.history/recent"
func (api API) getRecentSearchHistoryAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	searchHistory, err := model.GetSearchHistory(web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(searchHistory)
}

// GET "/api/search.history/pages/:count/:offset"
func (api API) getSearchHistoryByCountAndOffsetAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	count, err := r.RouteParamInt("count")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	searchHistory, err := model.GetSearchHistoryByCountAndOffset(count, offset, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(searchHistory)
}

// GET "/api/stats"
func (api API) getSiteStatsAction(r *web.Ctx) web.Result {
	stats, err := viewmodel.GetSiteStats(web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(stats)
}

// GET "/api/image.stats/:image_id"
func (api API) getImageStatsAction(r *web.Ctx) web.Result {
	imageUUID, err := r.RouteParam("image_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	image, err := model.GetImageByUUID(imageUUID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	stats, err := viewmodel.GetImageStats(image.ID, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(stats)
}

// GET "/api/session.user"
func (api API) getCurrentUserAction(r *web.Ctx) web.Result {
	session := r.Session()

	url, err := api.OAuth.OAuthURL()
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	cu := &viewmodel.CurrentUser{
		GoogleLoginURL: url,
		SlackLoginURL:  external.SlackAuthURL(api.Config),
	}
	if session == nil {
		return webutil.API(r).Result(cu)
	}

	cu.IsLoggedIn = true
	cu.SetFromUser(webutil.GetUser(session))
	return webutil.API(r).Result(cu)
}

// GET "/api/session/:key"
func (api API) getSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	value, hasValue := session.State[key]
	if !hasValue {
		return webutil.API(r).NotFound()
	}
	return webutil.API(r).Result(value)
}

// POST "/api/session/:key"
// PUT "/api/session/:key"
func (api API) setSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	session.State[key], err = r.PostBodyAsString()
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).OK()
}

// DELETE "/api/session/:key"
func (api API) deleteSessionKeyAction(r *web.Ctx) web.Result {
	session := r.Session()

	key, err := r.RouteParam("key")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}
	if _, hasKey := session.State[key]; !hasKey {
		return webutil.API(r).NotFound()
	}
	delete(session.State, key)
	return webutil.API(r).OK()
}

// GET "/api/jobs"
func (api API) getJobsStatusAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}
	status := cron.Default().Status()
	return webutil.API(r).Result(status)
}

// POST "/api/job/:job_id"
func (api API) runJobAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}
	jobID, err := r.RouteParam("job_id")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	err = cron.Default().RunJob(jobID)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}
	return webutil.API(r).OK()
}

func (api API) getErrorsAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if sessionUser != nil && !sessionUser.IsAdmin {
		return webutil.API(r).NotAuthorized()
	}

	limit, err := r.RouteParamInt("limit")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	offset, err := r.RouteParamInt("offset")
	if err != nil {
		return webutil.API(r).BadRequest(err)
	}

	errors, err := model.GetAllErrorsWithLimitAndOffset(limit, offset, web.Tx(r))
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).Result(errors)
}

// POST "/api/logout"
func (api API) logoutAction(r *web.Ctx) web.Result {
	session := r.Session()

	if session == nil {
		return webutil.API(r).OK()
	}

	err := r.Auth().Logout(r)
	if err != nil {
		return webutil.API(r).InternalError(err)
	}

	return webutil.API(r).OK()
}
