package main

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	var cfg config.Giffy
	if err := configutil.Read(&cfg); err != nil {
		logger.All().SyncFatalExit(err)
	}

	log := logger.NewFromConfig(&cfg.Logger)
	oauth := oauth.NewFromConfig(&cfg.GoogleAuth)

	if cfg.Web.IsSecure() {
		upgrader := web.NewHTTPSUpgraderFromConfig(&cfg.Upgrader).WithLogger(log)
		go func() {
			log.SyncFatal(upgrader.Start())
		}()
	}

	log.SyncFatal(server.New(log, oauth, &cfg).WithLogger(log).Start())
}
