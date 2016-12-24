package server

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-logger"
	request "github.com/blendlabs/go-request"
	"github.com/blendlabs/go-workqueue"
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
	"github.com/wcharczuk/go-web"

	// includes migrations
	_ "github.com/wcharczuk/giffy/database/migrations"
	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/jobs"
	"github.com/wcharczuk/giffy/server/model"
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
	return migration.Run(func(suite migration.Migration) error {
		suite.SetLogger(migration.NewLogger())
		return suite.Apply(spiffy.DefaultDb())
	})
}

// Init inits the web app.
func Init() *web.App {
	app := web.New()
	app.SetDiagnostics(logger.NewDiagnosticsAgentFromEnvironment())
	logger.SetDiagnostics(app.Diagnostics())
	app.Diagnostics().EventQueue().SetMaxWorkItems(1 << 18)
	app.SetAppName(AppName)
	app.SetPort(core.ConfigPort())

	app.Diagnostics().AddEventListener(logger.EventWebRequest, web.NewDiagnosticsRequestCompleteHandler(func(rc *web.RequestContext) {
		external.StatHatRequestTiming(rc.Elapsed())
	}))

	app.Diagnostics().AddEventListener(logger.EventFatalError, web.NewDiagnosticsErrorHandler(func(rc *web.RequestContext, err error) {
		external.StatHatError()
		model.DB().CreateInTx(model.NewError(err, rc.Request), rc.Tx())
	}))

	app.Diagnostics().AddEventListener(logger.EventError, web.NewDiagnosticsErrorHandler(func(rc *web.RequestContext, err error) {
		external.StatHatError()
		model.DB().CreateInTx(model.NewError(err, rc.Request), rc.Tx())
	}))

	app.Diagnostics().AddEventListener(
		request.EventFlagOutgoingResponse,
		request.NewEventFlagOutgoingResponseHandler(func(writer logger.Logger, ts logger.TimeSource, req *request.HTTPRequestMeta, res *request.HTTPResponseMeta, body []byte) {
			request.WriteOutgoingRequestResponse(writer, ts, req, res, body)
		}),
	)

	app.Register(new(controller.Index))
	app.Register(new(controller.API))
	app.Register(new(controller.Integrations))
	app.Register(new(controller.Auth))
	app.Register(new(controller.UploadImage))
	app.Register(new(controller.Chart))

	app.OnStart(func(app *web.App) error {
		err := core.DBInit()
		if err != nil {
			return err
		}

		if core.ConfigIsProduction() {
			app.
			ViewPaths = append(ViewPaths, "server/_views/header_prod.html")
		} else {
			ViewPaths = append(ViewPaths, "server/_views/header.html")
		}

		err = app.InitializeViewCache(ViewPaths...)
		if err != nil {
			return err
		}

		err = Migrate()
		if err != nil {
			return err
		}

		app.Diagnostics().EnableEvent(core.EventFlagSearch)
		app.Diagnostics().AddEventListener(core.EventFlagSearch, func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
			external.StatHatSearch()
			if len(state) > 0 {
				logger.WriteEventf(writer, ts, "Image Search", logger.ColorLightWhite, "query: %s", state[0].(*model.SearchHistory).SearchQuery)
				workQueue.Default().Enqueue(model.CreateObject, state[0])
			}
		})

		app.Diagnostics().EnableEvent(core.EventFlagModeration)
		app.Diagnostics().AddEventListener(core.EventFlagModeration, func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
			if len(state) > 0 {
				logger.WriteEventf(writer, ts, "Moderation", logger.ColorLightWhite, "verb: %s", state[0].(*model.Moderation).Verb)
				workQueue.Default().Enqueue(model.CreateObject, state[0])
			}
		})

		workQueue.Default().UseAsyncDispatch()
		workQueue.Default().Start()

		chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
		chronometer.Default().LoadJob(jobs.CleanTagValues{})
		chronometer.Default().LoadJob(jobs.FixContentRating{})
		chronometer.Default().LoadJob(jobs.FixImageSizes{})

		chronometer.Default().Start()
		return nil
	})

	return app
}
