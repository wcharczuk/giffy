package controller

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

// UploadImage is the controller responsible for image actions.
type UploadImage struct{}

func (ic UploadImage) uploadImageAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsModerator {
		return ctx.View.NotAuthorized()
	}

	return ctx.View.View("upload_image", nil)
}

func (ic UploadImage) uploadImageCompleteAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	if !session.User.IsModerator {
		return ctx.View.NotAuthorized()
	}

	var fileContents []byte
	var fileName string

	imageURL := ctx.Param("image_url")
	if len(imageURL) != 0 {

		refURL, err := url.Parse(imageURL)
		if err != nil {
			return ctx.View.BadRequest("`image_url` was malformed.")
		}

		res, err := external.NewRequest().
			AsGet().
			WithURL(refURL.String()).
			WithHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36").
			WithHeader("Cache-Control", "no-cache").
			FetchRawResponse()

		if err != nil {
			return ctx.View.InternalError(err)
		}

		if res.StatusCode != http.StatusOK {
			return ctx.View.BadRequest("Non 200 returned from `image_url` host.")
		}
		defer res.Body.Close()
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return ctx.View.InternalError(err)
		}

		fileName = path.Base(refURL.Path)
		fileContents = bytes
	} else {
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

		fileName = files[0].Filename
		fileContents = files[0].Contents
	}

	md5sum := model.ConvertMD5(md5.Sum(fileContents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return ctx.View.InternalError(err)
	}

	if !existing.IsZero() {
		return ctx.View.View("upload_image_complete", existing)
	}

	image, err := CreateImageFromFile(session.UserID, fileContents, fileName)
	if err != nil {
		return ctx.View.InternalError(err)
	}
	if image == nil {
		return ctx.View.InternalError(exception.New("Nil image returned from `createImageFromFile`."))
	}

	model.QueueModerationEntry(session.UserID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID)
	return ctx.View.View("upload_image_complete", image)
}

// CreateImageFromFile creates and uploads a new image.
func CreateImageFromFile(userID int64, fileContents []byte, fileName string) (*model.Image, error) {
	newImage, err := model.NewImageFromPostedFile(userID, fileContents, fileName)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(fileContents)
	remoteEntry, err := filecache.UploadFile(buf, filecache.FileType{Extension: newImage.Extension, MimeType: http.DetectContentType(fileContents)})
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

// Register registers the controllers routes.
func (ic UploadImage) Register(router *httprouter.Router) {
	router.GET("/images/upload", auth.ViewSessionRequiredAction(ic.uploadImageAction))
	router.POST("/images/upload", auth.ViewSessionRequiredAction(ic.uploadImageCompleteAction))
}
