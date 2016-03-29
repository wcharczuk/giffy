package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/julienschmidt/httprouter"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/jobs"
)

func indexAction(ctx *web.HTTPContext) web.ControllerResult {
	if core.ConfigIsProduction() {
		return ctx.Static("_client/dist/index.html")
	}
	return ctx.Static("_client/src/index.html")
}

func faviconAction(ctx *web.HTTPContext) web.ControllerResult {
	if core.ConfigIsProduction() {
		return ctx.Static("_client/dist/images/favicon.ico")
	}
	return ctx.Static("_client/src/images/favicon.ico")
}

// Init inits the app.
func Init() *httprouter.Router {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	web.SetLoggerStd()

	chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
	//chronometer.Default().LoadJob(jobs.FixImageSizes{})
	//chronometer.Default().LoadJob(jobs.CleanTagValues{})
	chronometer.Default().Start()

	paths := []string{
		"server/_views/footer.html",
		"server/_views/not_found.html",
		"server/_views/error.html",
		"server/_views/bad_request.html",
		"server/_views/not_authorized.html",
		"server/_views/login_complete.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	}

	if core.ConfigIsProduction() {
		paths = append(paths, "server/_views/header_prod.html")
	} else {
		paths = append(paths, "server/_views/header.html")
	}

	web.InitViewCache(paths...)

	router := httprouter.New()

	new(controller.API).Register(router)
	new(controller.Auth).Register(router)
	new(controller.UploadImage).Register(router)

	router.GET("/", web.ActionHandler(indexAction))
	router.GET("/favicon.ico", web.ActionHandler(faviconAction))
	if core.ConfigIsProduction() {
		router.ServeFiles("/static/*filepath", http.Dir("_client/dist"))
	} else {
		router.ServeFiles("/bower/*filepath", http.Dir("_client/bower"))
		router.ServeFiles("/static/*filepath", http.Dir("_client/src"))
	}

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
