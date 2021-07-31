package controller

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"path"

	exception "github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/r2"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
)

// UploadImage is the controller responsible for image actions.
type UploadImage struct {
	Log    logger.Log
	Config *config.Giffy
	Model  *model.Manager
	Files  *filemanager.FileManager
}

// Register registers the controllers routes.
func (ic UploadImage) Register(app *web.App) {
	app.GET("/images/upload", ic.uploadImageAction, web.SessionRequired, web.ViewProviderAsDefault)
	app.POST("/images/upload", ic.uploadImageCompleteAction, web.SessionRequired, web.ViewProviderAsDefault)
}

func (ic UploadImage) uploadImageAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if !sessionUser.IsModerator {
		return r.Views.NotAuthorized()
	}

	return r.Views.View("upload_image", nil)
}

func (ic UploadImage) uploadImageCompleteAction(r *web.Ctx) web.Result {
	sessionUser := GetUser(r.Session)
	if !sessionUser.IsModerator {
		return r.Views.NotAuthorized()
	}

	var fileContents []byte
	var fileName string
	if imageURL := web.StringValue(r.Param("image_url")); imageURL != "" {
		refURL, err := url.Parse(imageURL)
		if err != nil {
			return r.Views.BadRequest(fmt.Errorf("`image_url` was malformed"))
		}
		contents, res, err := r2.New(refURL.String(),
			r2.OptLog(ic.Log),
			r2.OptHeaderValue("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36"),
			r2.OptHeaderValue("Cache-Control", "no-cache"),
		).Bytes()
		if err != nil {
			return r.Views.InternalError(err)
		}
		if res.StatusCode != http.StatusOK {
			return r.Views.BadRequest(fmt.Errorf("non 200 returned from `image_url` host"))
		}
		fileName = path.Base(refURL.Path)
		fileContents = contents
	} else {
		files, err := webutil.PostedFiles(
			r.Request,
			webutil.OptPostedFilesParseMultipartForm(true),
			webutil.OptPostedFilesParseForm(false),
		)
		if err != nil {
			return r.Views.BadRequest(fmt.Errorf("problem reading posted file: %v", err))
		}
		if len(files) == 0 {
			return r.Views.BadRequest(fmt.Errorf("no files posted"))
		}
		if len(files) > 1 {
			return r.Views.BadRequest(fmt.Errorf("too many files posted"))
		}
		fileName = files[0].FileName
		fileContents = files[0].Contents
	}

	md5sum := model.ConvertMD5(md5.Sum(fileContents))
	existing, err := ic.Model.GetImageByMD5(r.Context(), md5sum)
	if err != nil {
		return r.Views.InternalError(err)
	}

	if !existing.IsZero() {
		return r.Views.View("upload_image_complete", existing)
	}

	image, err := CreateImageFromFile(r.Context(), ic.Model, sessionUser.ID, !sessionUser.IsAdmin, fileContents, fileName, ic.Files)
	if err != nil {
		return r.Views.InternalError(err)
	}
	if image == nil {
		return r.Views.InternalError(exception.New("image returned from `createImageFromFile` was unset"))
	}

	logger.MaybeTrigger(r.Context(), ic.Log, model.NewModeration(sessionUser.ID, model.ModerationVerbCreate, model.ModerationObjectImage, image.UUID))
	return r.Views.View("upload_image_complete", image)
}

// CreateImageFromFile creates and uploads a new image.
func CreateImageFromFile(ctx context.Context, mgr *model.Manager, userID int64, shouldValidate bool, fileContents []byte, fileName string, fm *filemanager.FileManager) (*model.Image, error) {
	newImage, err := model.NewImageFromPostedFile(userID, shouldValidate, fileContents, fileName)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(fileContents)
	remoteEntry, err := fm.UploadFile(buf, filemanager.FileType{Extension: newImage.Extension, MimeType: http.DetectContentType(fileContents)})
	if err != nil {
		return nil, err
	}

	newImage.S3Bucket = remoteEntry.Bucket
	newImage.S3Key = remoteEntry.Key

	err = mgr.Invoke(ctx).Create(newImage)
	if err != nil {
		return nil, err
	}

	return newImage, nil
}
