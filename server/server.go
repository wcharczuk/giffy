package server

import (
	"fmt"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util"
	"github.com/stathat/go"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
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
	app.InitViewCache(paths...)
	app.SetLogger(web.NewStandardOutputLogger())

	app.OnRequestComplete(func(r *web.RequestContext) {
		statHatToken := core.ConfigStathatToken()
		requestTimingBucket := fmt.Sprintf("request_timing_%s", core.ConfigEnvironment())
		requestElapsed := r.Elapsed()
		stathat.PostEZValue(requestTimingBucket, statHatToken, float64(requestElapsed/time.Millisecond))
	})

	app.Register(new(controller.Index))
	app.Register(new(controller.API))
	app.Register(new(controller.Auth))
	app.Register(new(controller.UploadImage))

	return app
}
