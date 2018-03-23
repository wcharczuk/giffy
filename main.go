package main

import (
	"fmt"
	"os"

	google "github.com/blendlabs/go-google-oauth"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/configutil"
	web "github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	var cfg config.Giffy
	err := configutil.Read(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	log := logger.NewFromConfig(&cfg.Logger)
	oauth := google.NewFromConfig(&cfg.GoogleAuth)

	if cfg.Web.IsSecure() {
		upgrader := web.NewHTTPSUpgraderFromConfig(&cfg.Upgrader).WithLogger(log)
		go func() {
			log.SyncFatal(upgrader.Start())
		}()
	}

	log.SyncFatal(server.New(log, oauth, &cfg).WithLogger(log).Start())
}
