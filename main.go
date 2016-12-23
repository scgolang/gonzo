package main

import (
	"context"
	"log"
)

const (
	// ApplicationName is the name of the application.
	ApplicationName = "gonzo"

	// WelcomeMessage is sent to clients in the announce reply.
	WelcomeMessage = `welcome to gonzo`
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	app, err := NewApp(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Wait(); err != nil {
		log.Fatal(err)
	}
}
