package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/filecache"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

const (
	// OAuthProviderGoogle is the only auth provider we use right now.
	OAuthProviderGoogle = "google"
)

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

func indexAction(ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("server/_static/index.html")
}

// Init inits the app.
func Init() *httprouter.Router {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	)

	router := httprouter.New()

	new(APIController).Register(router)
	new(AuthController).Register(router)
	new(ImageController).Register(router)

	router.GET("/", web.ActionHandler(indexAction))
	router.ServeFiles("/static/*filepath", http.Dir("server/_static"))

	router.NotFound = web.NotFoundHandler
	router.PanicHandler = web.PanicHandler

	return router
}

// Start starts the app.
func Start(router *httprouter.Router) {
	bindAddr := fmt.Sprintf(":%s", core.ConfigPort())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	web.Logf("Giffy Server Started, listening on %s", bindAddr)
	log.Fatal(server.ListenAndServe())
}
