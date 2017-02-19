package controller

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/filecache"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/webutil"
)

// UploadImage is the controller responsible for image actions.
type UploadImage struct{}

func (ic UploadImage) uploadImageAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if !sessionUser.IsModerator {
		return r.View().NotAuthorized()
	}

	return r.View().View("upload_image", nil)
}

func (ic UploadImage) uploadImageCompleteAction(r *web.Ctx) web.Result {
	sessionUser := webutil.GetUser(r.Session())
	if !sessionUser.IsModerator {
		return r.View().NotAuthorized()
	}

	var fileContents []byte
	var fileName string

	imageURL := r.Param("image_url")
	if len(imageURL) != 0 {
		refURL, err := url.Parse(imageURL)
		if err != nil {
			return r.View().BadRequest("`image_url` was malformed.")
		}

		res, err := external.NewRequest().
			AsGet().
			WithURL(refURL.String()).
			WithHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36").
			WithHeader("Cache-Control", "no-cache").
			FetchRawResponse()

		if err != nil {
			return r.View().InternalError(err)
		}

		if res.StatusCode != http.StatusOK {
			return r.View().BadRequest("Non 200 returned from `image_url` host.")
		}
		defer res.Body.Close()
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return r.View().InternalError(err)
		}

		fileName = path.Base(refURL.Path)
		fileContents = bytes
	} else {
		files, filesErr := r.PostedFiles()
		if filesErr != nil {
			return r.View().BadRequest(fmt.Sprintf("Problem reading posted file: %v", filesErr))
		}

		if len(files) == 0 {
			return r.View().BadRequest("No files posted.")
		}

		if len(files) > 1 {
			return r.View().BadRequest("Too many files posted.")
		}

		fileName = files[0].FileName
		fileContents = files[0].Contents
	}

	md5sum := model.ConvertMD5(md5.Sum(fileContents))
	existing, err := model.GetImageByMD5(md5sum, nil)
	if err != nil {
		return r.View().InternalError(err)
	}

	if !existing.IsZero() {
		return r.View().View("upload_image_complete", existing)
	}

	image, err := CreateImageFromFile(sessionUser.ID, !sessionUser.IsAdmin, fileContents, fileName, r.Tx())
	if err != nil {
		return r.View().InternalError(err)
	}
	if image == nil {
		return r.View().InternalError(exception.New("Nil image returned from `createImageFromFile`."))
	}

	r.Diagnostics().OnEvent(core.EventFlagModeration, model.NewModeration(sessionUser.ID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID))
	return r.View().View("upload_image_complete", image)
}

// CreateImageFromFile creates and uploads a new image.
func CreateImageFromFile(userID int64, shouldValidate bool, fileContents []byte, fileName string, tx *sql.Tx) (*model.Image, error) {
	newImage, err := model.NewImageFromPostedFile(userID, shouldValidate, fileContents, fileName)
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

	err = spiffy.DefaultDb().CreateInTx(newImage, tx)
	if err != nil {
		return nil, err
	}

	return newImage, nil
}

// Register registers the controllers routes.
func (ic UploadImage) Register(app *web.App) {
	app.GET("/images/upload", ic.uploadImageAction, web.SessionRequired, web.ViewProviderAsDefault)
	app.POST("/images/upload", ic.uploadImageCompleteAction, web.SessionRequired, web.ViewProviderAsDefault)
}
