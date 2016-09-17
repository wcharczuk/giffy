package server

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-workqueue"
	"github.com/wcharczuk/go-web"

	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/jobs"
	"github.com/wcharczuk/giffy/server/migrate"
)

const (
	// AppName is the name of the app.
	AppName = "giffy"
)

var (
	// ViewPaths are paths to load into the view cache.
	ViewPaths = []string{
		"server/_views/footer.html",
		"server/_views/not_found.html",
		"server/_views/error.html",
		"server/_views/bad_request.html",
		"server/_views/not_authorized.html",
		"server/_views/login_complete.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	}
)

// Migrate migrates the db
func Migrate() error {
	migrate.Register()
	err := migrate.Run()
	if err != nil {
		return err
	}
	return nil
}

// Init inits the web app.
func Init() (*web.App, error) {
	diagnostics, err := logger.NewDiagnosticsAgentFromEnvironment()
	if err != nil {
		return nil, err
	}
	app := web.New()
	app.SetDiagnostics(diagnostics)
	app.SetAppName(AppName)
	app.SetPort(core.ConfigPort())

	app.Diagnostics().AddEventListener(logger.EventRequestComplete, func(wr logger.Logger, ts logger.TimeSource, eventFlag uint64, state ...interface{}) {
		context := state[0].(*web.RequestContext)
		external.StatHatRequestTiming(context.Elapsed())
	})

	app.Diagnostics().AddEventListener(logger.EventError, func(wr logger.Logger, ts logger.TimeSource, eventFlag uint64, state ...interface{}) {
		external.StatHatError()
	})

	app.Register(new(controller.Index))
	app.Register(new(controller.API))
	app.Register(new(controller.Integrations))
	app.Register(new(controller.Auth))
	app.Register(new(controller.UploadImage))

	app.OnStart(func(app *web.App) error {
		err := core.DBInit()
		if err != nil {
			return err
		}

		if core.ConfigIsProduction() {
			ViewPaths = append(ViewPaths, "server/_views/header_prod.html")
		} else {
			ViewPaths = append(ViewPaths, "server/_views/header.html")
		}

		err = app.InitViewCache(ViewPaths...)
		if err != nil {
			return err
		}

		err = Migrate()
		if err != nil {
			return err
		}

		workQueue.Default().UseAsyncDispatch()
		workQueue.Default().Start()

		chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
		chronometer.Default().Start()
		return nil
	})

	return app, nil
}
