package server

import (
	"context"
	"net/http"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

	// includes migrations
	migrations "github.com/wcharczuk/giffy/db/migrations"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/controller"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/filemanager"
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

// New returns a new server instance.
func New(cfg *config.Giffy) (*web.App, error) {
	log := logger.NewFromConfig(&cfg.Logger)
	log.Enable(core.FlagSearch, core.FlagModeration)

	conn, err := db.NewFromConfig(&cfg.DB)
	if err != nil {
		return nil, err
	}
	if err := conn.Open(); err != nil {
		return nil, err
	}
	conn.WithLogger(log)

	log.Infof("using database: %s", conn.Config().CreateDSN())

	mgr := &model.Manager{DB: conn}

	oauthMgr, err := oauth.NewFromConfig(&cfg.GoogleAuth)
	if err != nil {
		return nil, err
	}

	app := web.NewFromConfig(&cfg.Web).WithLogger(log)
	if env.Env().Has("CURRENT_REF") {
		app.WithDefaultHeader("X-Server-Version", env.Env().String("CURRENT_REF"))
	}

	log.Listen(logger.Fatal, "error-writer", logger.NewErrorEventListener(func(ev *logger.ErrorEvent) {
		if req, isReq := ev.State().(*http.Request); isReq {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err(), req))
		} else {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err(), nil))
		}
	}))
	log.Listen(logger.Error, "error-writer", logger.NewErrorEventListener(func(ev *logger.ErrorEvent) {
		if req, isReq := ev.State().(*http.Request); isReq {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err(), req))
		} else {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err(), nil))
		}
	}))
	log.Listen(core.FlagSearch, "event-writer", func(e logger.Event) {
		if typed, ok := e.(db.DatabaseMapped); ok {
			logger.MaybeError(log, mgr.Invoke(context.Background()).Create(typed))
		}
	})
	log.Listen(core.FlagModeration, "event-writer", func(e logger.Event) {
		if typed, ok := e.(db.DatabaseMapped); ok {
			logger.MaybeError(log, mgr.Invoke(context.Background()).Create(typed))
		}
	})

	if cfg.IsProduction() {
		app.Views().AddPaths("server/_views/header_prod.html")
	} else {
		app.Views().AddPaths("server/_views/header.html")
	}

	app.Views().AddPaths(ViewPaths...)

	fm := filemanager.New(cfg.GetS3Bucket(), &cfg.Aws)

	app.Register(controller.Index{Model: mgr, Config: cfg})
	app.Register(controller.APIs{Model: mgr, Config: cfg, Files: fm, OAuth: oauthMgr})
	app.Register(controller.Integrations{Model: mgr, Config: cfg})
	app.Register(controller.Auth{Model: mgr, Config: cfg, OAuth: oauthMgr})
	app.Register(controller.UploadImage{Model: mgr, Config: cfg, Files: fm})
	app.Register(controller.Chart{Model: mgr, Config: cfg})

	if migrations.Migrations().Apply(conn); err != nil {
		return nil, err
	}

	cron.Default().LoadJob(jobs.DeleteOrphanedTags{})
	cron.Default().LoadJob(jobs.CleanTagValues{Model: mgr})
	cron.Default().LoadJob(jobs.FixContentRating{})
	cron.Default().Start()

	return app, nil
}
