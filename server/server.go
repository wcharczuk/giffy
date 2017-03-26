package server

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/go-workqueue"
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"

	// includes migrations
	_ "github.com/wcharczuk/giffy/database/migrations"
	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/jobs"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/webutil"
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
		suite.SetLogger(migration.NewLoggerFromAgent(logger.NewFromEnvironment()))
		return suite.Apply(spiffy.DB())
	})
}

// New returns a new server instance.
func New() *web.App {
	app := web.New()
	app.SetLogger(logger.NewFromEnvironment())
	logger.SetDefault(app.Logger())
	app.SetName(AppName)
	app.SetPort(core.ConfigPort())

	app.Logger().DisableEvent(logger.EventWebRequestPostBody)
	app.Logger().DisableEvent(logger.EventWebResponse)

	app.Logger().AddEventListener(logger.EventWebRequest, web.NewDiagnosticsRequestCompleteHandler(func(rc *web.Ctx) {
		external.StatHatRequestTiming(rc.Elapsed())
	}))

	app.Logger().AddEventListener(logger.EventFatalError, web.NewDiagnosticsErrorHandler(func(rc *web.Ctx, err error) {
		external.StatHatError()
		model.DB().CreateInTx(model.NewError(err, rc.Request), rc.Tx())
	}))

	app.Logger().AddEventListener(logger.EventError, web.NewDiagnosticsErrorHandler(func(rc *web.Ctx, err error) {
		external.StatHatError()
		model.DB().CreateInTx(model.NewError(err, rc.Request), rc.Tx())
	}))

	app.Logger().AddEventListener(
		request.Event,
		request.NewOutgoingListener(func(wr logger.Logger, ts logger.TimeSource, req *request.HTTPRequestMeta) {
			request.WriteOutgoingRequest(wr, ts, req)
		}),
	)

	if app.Logger().IsEnabled(logger.EventDebug) {
		app.Logger().EnableEvent(spiffy.EventFlagQuery)
		app.Logger().AddEventListener(
			spiffy.EventFlagQuery,
			spiffy.NewPrintStatementListener(),
		)

		app.Logger().EnableEvent(spiffy.EventFlagExecute)
		app.Logger().AddEventListener(
			spiffy.EventFlagExecute,
			spiffy.NewPrintStatementListener(),
		)
	}

	app.Logger().EnableEvent(core.EventFlagSearch)
	app.Logger().AddEventListener(core.EventFlagSearch, func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		external.StatHatSearch()
		if len(state) > 0 {
			logger.WriteEventf(writer, ts, "Image Search", logger.ColorLightWhite, "query: %s", state[0].(*model.SearchHistory).SearchQuery)
			workqueue.Default().Enqueue(model.CreateObject, state[0])
		}
	})

	app.Logger().EnableEvent(core.EventFlagModeration)
	app.Logger().AddEventListener(core.EventFlagModeration, func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		if len(state) > 0 {
			logger.WriteEventf(writer, ts, "Moderation", logger.ColorLightWhite, "verb: %s", state[0].(*model.Moderation).Verb)
			workqueue.Default().Enqueue(model.CreateObject, state[0])
		}
	})

	if core.ConfigIsProduction() {
		app.ViewCache().AddPaths("server/_views/header_prod.html")
	} else {
		app.ViewCache().AddPaths("server/_views/header.html")
	}

	webutil.LiveReloads(app)
	webutil.BaseURL(app)
	webutil.SecureCookies(app)

	app.Auth().SetSessionParamName("giffy")

	app.ViewCache().AddPaths(ViewPaths...)

	app.Register(new(controller.Index))
	app.Register(new(controller.API))
	app.Register(new(controller.Integrations))
	app.Register(new(controller.Auth))
	app.Register(new(controller.UploadImage))
	app.Register(new(controller.Chart))

	app.OnStart(func(a *web.App) error {
		err := core.DBInit()
		if err != nil {
			return err
		}

		spiffy.DB().SetLogger(a.Logger())

		err = Migrate()
		if err != nil {
			return err
		}

		workqueue.Default().Start()

		chronometer.Default().LoadJob(jobs.DeleteOrphanedTags{})
		chronometer.Default().LoadJob(jobs.CleanTagValues{})
		chronometer.Default().LoadJob(jobs.FixContentRating{})
		chronometer.Default().LoadJob(jobs.FixImageSizes{})
		chronometer.Default().Start()

		return nil
	})

	return app
}
