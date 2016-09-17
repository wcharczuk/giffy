package main

import (
	"log"

	"github.com/wcharczuk/giffy/server"
)

func main() {
	app, err := server.Init()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(app.Start())
}
