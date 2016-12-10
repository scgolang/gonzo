package main

import (
	"context"
	"log"
	"os"
	"path"
)

var (
	Home = path.Join(os.Getenv("HOME"), "gonzo-sessions")
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

	app.Go(app.ServeOSC)

	if err := app.Wait(); err != nil {
		log.Fatal(err)
	}
}
