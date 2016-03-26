package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"

	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/jobs"
)

func indexAction(ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("_static/index.html")
}

func faviconAction(ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("_static/images/favicon.ico")
}

// Init inits the app.
func Init() *httprouter.Router {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
	chronometer.Default().LoadJob(jobs.FixImageSizes{})
	chronometer.Default().Start()

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/not_found.html",
		"server/_views/error.html",
		"server/_views/bad_request.html",
		"server/_views/not_authorized.html",
		"server/_views/login_complete.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	)

	router := httprouter.New()

	new(controller.API).Register(router)
	new(controller.Auth).Register(router)
	new(controller.UploadImage).Register(router)

	router.GET("/", web.ActionHandler(indexAction))
	router.GET("/favicon.ico", web.ActionHandler(faviconAction))
	router.ServeFiles("/static/*filepath", http.Dir("_static"))
	router.ServeFiles("/_bower/*filepath", http.Dir("_bower"))

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
