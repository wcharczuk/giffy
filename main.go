package main

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"

	"github.com/wcharczuk/giffy/server"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	var cfg config.Giffy
	if _, err := configutil.Read(&cfg); !configutil.IsIgnored(err) {
		logger.FatalExit(err)
	}

	app, err := server.New(&cfg)
	if err != nil {
		logger.FatalExit(err)
	}

	if err = graceful.Shutdown(app); err != nil {
		logger.FatalExit(err)
	}
}
