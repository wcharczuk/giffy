package main

import (
	"log"

	"github.com/blendlabs/go-util/configutil"
	"github.com/wcharczuk/giffy/server"
	"github.com/wcharczuk/giffy/server/config"
)

func main() {
	var cfg config.Giffy
	err := configutil.Read(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.New(&cfg).Start())
}
