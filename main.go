package main

import (
	"log"

	_ "github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/route"
)

func main() {
	app := route.InitRoutes()
	log.Fatal(app.Listen("localhost:8080"))
}
