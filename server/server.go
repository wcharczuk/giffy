package server

import (
	"context"
	"net/http"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

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
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/not_found.html",
		"server/_views/error.html",
		"server/_views/bad_request.html",
		"server/_views/not_authorized.html",
		"server/_views/login_complete.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/privacy.html",
		"server/_views/tos.html",
	}
)

// New returns a new server instance.
func New(cfg *config.Giffy) (*web.App, error) {
	log := logger.MustNew(
		logger.OptConfig(cfg.Logger),
	)
	log.Enable(core.FlagSearch, core.FlagModeration)
	log.Disable(db.QueryStartFlag)

	conn, err := db.New(
		db.OptConfig(cfg.DB),
		db.OptLog(log),
	)
	if err != nil {
		return nil, err
	}
	if err := conn.Open(); err != nil {
		return nil, err
	}

	log.Infof("using database: %s", conn.Config.CreateLoggingDSN())
	log.Infof("using aws access key: %s", cfg.Aws.AccessKeyID)
	log.Infof("using aws region: %s", cfg.Aws.RegionOrDefault())
	log.Infof("using admin user email: %s", cfg.AdminUserEmail)
	log.Infof("using cloudfront dns: %s", cfg.CloudFrontDNS)
	log.Infof("using s3 bucket: %s", cfg.S3Bucket)

	mgr := &model.Manager{BaseManager: dbutil.NewBaseManager(conn)}

	oauthMgr, err := oauth.New(
		oauth.OptConfig(cfg.GoogleAuth),
	)
	if err != nil {
		return nil, err
	}

	app := web.MustNew(web.OptConfig(cfg.Web), web.OptLog(log))
	if cfg.Meta.Version != "" {
		app.BaseHeaders.Add("X-Server-Version", cfg.Meta.Version)
	} else if cfg.Meta.GitRef != "" {
		app.BaseHeaders.Add("X-Server-Version", cfg.Meta.GitRef)
	}

	log.Listen(logger.Fatal, "error-writer", logger.NewErrorEventListener(func(_ context.Context, ev logger.ErrorEvent) {
		if req, isReq := ev.State.(*http.Request); isReq {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err, req))
		} else {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err, nil))
		}
	}))
	log.Listen(logger.Error, "error-writer", logger.NewErrorEventListener(func(_ context.Context, ev logger.ErrorEvent) {
		if req, isReq := ev.State.(*http.Request); isReq {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err, req))
		} else {
			mgr.Invoke(context.Background()).Create(model.NewError(ev.Err, nil))
		}
	}))
	log.Listen(core.FlagSearch, "event-writer", func(_ context.Context, e logger.Event) {
		if typed, ok := e.(db.DatabaseMapped); ok {
			logger.MaybeError(log, mgr.Invoke(context.Background()).Create(typed))
		}
	})
	log.Listen(core.FlagModeration, "event-writer", func(_ context.Context, e logger.Event) {
		if typed, ok := e.(db.DatabaseMapped); ok {
			logger.MaybeError(log, mgr.Invoke(context.Background()).Create(typed))
		}
	})

	app.Views.AddPaths(ViewPaths...)

	fm := filemanager.New(cfg.S3Bucket, cfg.Aws)

	app.Register(controller.Index{Log: log, Model: mgr, Config: cfg})
	app.Register(controller.APIs{Log: log, Model: mgr, Config: cfg, Files: fm, OAuth: oauthMgr})
	app.Register(controller.Integrations{Log: log, Model: mgr, Config: cfg})
	app.Register(controller.Auth{Log: log, Model: mgr, Config: cfg, OAuth: oauthMgr})
	app.Register(controller.UploadImage{Log: log, Model: mgr, Config: cfg, Files: fm})
	app.Register(controller.Chart{Model: mgr, Config: cfg})

	if model.Migrations(cfg).Apply(context.Background(), conn); err != nil {
		return nil, err
	}

	cron.Default().Log = log.WithPath("jobs")
	cron.Default().LoadJobs(jobs.DeleteOrphanedTags{Model: mgr})
	cron.Default().LoadJobs(jobs.CleanTagValues{Model: mgr})
	cron.Default().LoadJobs(jobs.FixContentRating{Model: mgr})
	cron.Default().StartAsync()

	return app, nil
}
