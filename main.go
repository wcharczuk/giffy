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

	err = env.Env().ReadInto(&cfg)
	if err != nil {
		log.SyncWarning(err)
	}

	if cfg.Web.TLS.HasKeyPair() || cfg.Web.BaseURLIsSecureScheme() {

	}

	server.New(&cfg).WithLogger(log).Start()
}
