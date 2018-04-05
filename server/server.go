package server

import (
	"net/http"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/spiffy"
	"github.com/blend/go-sdk/spiffy/migration"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/workqueue"

	// includes migrations
	_ "github.com/wcharczuk/giffy/database/migrations"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/filemanager"
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
	return migration.Default().Apply(model.DB())
}

// New returns a new server instance.
func New(log *logger.Logger, oauth *oauth.Manager, cfg *config.Giffy) *web.App {
	app := web.NewFromConfig(&cfg.Web).WithLogger(log)

	app.Logger().Listen(logger.Fatal, "error-writer", logger.NewErrorEventListener(func(ev *logger.ErrorEvent) {
		if req, isReq := ev.State().(*http.Request); isReq {
			model.DB().Create(model.NewError(ev.Err(), req))
		}
	}))
	app.Logger().Listen(logger.Error, "error-writer", logger.NewErrorEventListener(func(ev *logger.ErrorEvent) {
		if req, isReq := ev.State().(*http.Request); isReq {
			model.DB().Create(model.NewError(ev.Err(), req))
		}
	}))

	app.Logger().Enable(core.FlagSearch, core.FlagModeration)
	app.Logger().Listen(core.FlagSearch, "event-writer", func(e logger.Event) {
		workqueue.Default().Enqueue(model.CreateObject, e)
	})
	app.Logger().Listen(core.FlagModeration, "event-writer", func(e logger.Event) {
		workqueue.Default().Enqueue(model.CreateObject, e)
	})

	if cfg.IsProduction() {
		app.Views().AddPaths("server/_views/header_prod.html")
	} else {
		app.Views().AddPaths("server/_views/header.html")
	}

	webutil.LiveReloads(app)
	webutil.BaseURL(app)
	webutil.SecureCookies(app)

	app.Auth().SetCookieName("giffy")
	app.Views().AddPaths(ViewPaths...)

	fm := filemanager.New(cfg.GetS3Bucket(), &cfg.Aws)

	app.Register(controller.Index{Config: cfg})
	app.Register(controller.API{Config: cfg, Files: fm, OAuth: oauth})
	app.Register(controller.Integrations{Config: cfg})
	app.Register(controller.Auth{Config: cfg, OAuth: oauth})
	app.Register(controller.UploadImage{Config: cfg, Files: fm})
	app.Register(controller.Chart{Config: cfg})

	app.OnStart(func(a *web.App) error {
		err := spiffy.OpenDefault(spiffy.NewFromConfig(&cfg.DB))
		if err != nil {
			return err
		}

		model.DB().WithLogger(a.Logger())

		err = Migrate()
		if err != nil {
			return err
		}

		workqueue.Default().Start()

		cron.Default().LoadJob(jobs.DeleteOrphanedTags{})
		cron.Default().LoadJob(jobs.CleanTagValues{})
		cron.Default().LoadJob(jobs.FixContentRating{})
		cron.Default().Start()

		return nil
	})

	return app
}
