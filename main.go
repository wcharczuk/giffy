package main

import (
	"log"

	"github.com/wcharczuk/giffy/server"
)

func main() {
	log.Fatal(server.Init().Start())
}
