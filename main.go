package main

import (
	"fmt"
	"os"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/configutil"
	"github.com/blendlabs/go-util/env"
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

	if cfg.Web.IsSecure() {
		upgrader := web.NewHTTPSUpgraderFromConfig(&cfg.Upgrader)
		go log.SyncFatal(upgrader.Start()))))
	}

	server.New(&cfg).WithLogger(log).Start()
}
