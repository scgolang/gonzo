package main

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/scgolang/nsm"
)

const (
	// ApplicationName is the name of the application.
	ApplicationName = "gonzo"

	// WelcomeMessage is sent to clients in the announce reply.
	WelcomeMessage = `welcome to gonzo`
)

var (
	Home         = path.Join(os.Getenv("HOME"), "gonzo-sessions")
	Capabilities = nsm.Capabilities{nsm.CapServerControl}
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
