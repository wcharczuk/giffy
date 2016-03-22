package controller

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

// UploadImage is the controller responsible for image actions.
type UploadImage struct{}

func (ic UploadImage) uploadImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.View.View("upload_image", nil)
}

func (ic UploadImage) uploadImageCompleteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	files, filesErr := ctx.PostedFiles()
	if filesErr != nil {
		return ctx.View.BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
	}

	if len(files) == 0 {
		return ctx.View.BadRequest("No files posted.")
	}

	if len(files) > 1 {
		return ctx.View.BadRequest("Too many files posted.")
	}

	postedFile := files[0]

	md5sum := model.ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	if !existing.IsZero() {
		return ctx.View.View("upload_image_complete", existing)
	}

	image, err := CreateImageFromFile(session.UserID, postedFile)
	if err != nil {
		return ctx.View.InternalError(err)
	}
	if image == nil {
		return ctx.View.InternalError(exception.New("Nil image returned from `createImageFromFile`."))
	}
	tagValue := strings.ToLower(ctx.Param("tag_value"))

	existingTag, err := model.GetTagByValue(tagValue, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	var tagID int64
	var tagUUID string
	if existingTag.IsZero() {
		newTag := model.NewTag(session.UserID, tagValue)
		err = spiffy.DefaultDb().Create(newTag)
		if err != nil {
			return ctx.View.InternalError(err)
		}
		tagID = newTag.ID
		tagUUID = newTag.UUID

		err = model.UpdateImageDisplayName(image.ID, tagValue, nil)
		if err != nil {
			return ctx.View.InternalError(err)
		}
	} else {
		tagID = existingTag.ID
		tagUUID = existingTag.UUID
	}

	// automatically vote for the tag <==> image
	didCreateLink, err := model.CreateOrIncrementVote(session.UserID, image.ID, tagID, true, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
	if didCreateLink {
		model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectLink, image.UUID, tagUUID)
	}
	return ctx.View.View("upload_image_complete", image)
}

// Register registers the controllers routes.
func (ic UploadImage) Register(router *httprouter.Router) {
	router.GET("/images/upload", auth.ViewSessionRequiredAction(ic.uploadImageAction))
	router.POST("/images/upload", auth.ViewSessionRequiredAction(ic.uploadImageCompleteAction))
}

// CreateImageFromFile creates and uploads a new image.
func CreateImageFromFile(userID int64, file web.PostedFile) (*model.Image, error) {
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
