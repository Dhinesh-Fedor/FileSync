package main

import (
	"log"

	"FileSync/internal/server"
)

func main() {
	app, err := server.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Listen(""); err != nil {
		log.Fatal(err)
	}
}
