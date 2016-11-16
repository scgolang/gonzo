package main

import (
	"log"
	"os"
	"path"
)

var (
	Home = path.Join(os.Getenv("HOME"), "nsm-sessions")
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	app, err := NewApp(config)
	if err != nil {
		log.Fatal(err)
	}

	app.Go(app.ServeOSC)

	if err := app.Wait(); err != nil {
		log.Fatal(err)
	}
}
