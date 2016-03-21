package server

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

// ImageController is the controller responsible for image actions.
type ImageController struct{}

func (ic ImageController) uploadImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.View("upload_image", nil)
}

func (ic ImageController) uploadImageCompleteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
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

	image, err := CreateImageFromFile(session.UserID, postedFile)
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
	var tagUUID string
	if existingTag.IsZero() {
		newTag := model.NewTag(session.UserID, tagValue)
		err = spiffy.DefaultDb().Create(newTag)
		if err != nil {
			return ctx.InternalError(err)
		}
		tagID = newTag.ID
		tagUUID = newTag.UUID

		err = model.UpdateImageDisplayName(image.ID, tagValue, nil)
		if err != nil {
			return ctx.InternalError(err)
		}
	} else {
		tagID = existingTag.ID
		tagUUID = existingTag.UUID
	}

	// automatically vote for the tag <==> image
	didCreate, err := model.CreateOrIncrementVote(session.UserID, image.ID, tagID, true, nil)
	if err != nil {
		return ctx.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
	if didCreate {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectLink, image.UUID, tagUUID)
	}
	return ctx.View("upload_image_complete", image)
}

// Register registers the controllers routes.
func (ic ImageController) Register(router *httprouter.Router) {
	router.GET("/images/upload", web.ActionHandler(auth.SessionRequiredAction(ic.uploadImageAction)))
	router.POST("/images/upload", web.ActionHandler(auth.SessionRequiredAction(ic.uploadImageCompleteAction)))
}
