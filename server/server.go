package server

import (
	"log"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/jobs"
)

// Init inits the app.
func Init() *web.App {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
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

	app := web.New()
	app.SetName("giffy")
	app.SetPort(core.ConfigPort())
	viewCacheErr := app.InitViewCache(paths...)
	if viewCacheErr != nil {
		log.Fatal(viewCacheErr)
	}
	if !core.ConfigIsProduction() {
		app.SetLogger(web.NewStandardOutputLogger())
	}

	app.RequestCompleteHandler(func(r *web.RequestContext) {
		external.StatHatRequestTiming(r.Elapsed())
	})

	app.RequestErrorHandler(func(r *web.RequestContext, err interface{}) {
		external.StatHatError()
	})

	app.Register(new(controller.Index))
	app.Register(new(controller.API))
	app.Register(new(controller.Integrations))
	app.Register(new(controller.Auth))
	app.Register(new(controller.UploadImage))

	return app
}
