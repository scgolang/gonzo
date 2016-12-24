package main

import (
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Add starts a new client program.
func (app *App) Add(msg osc.Message) error {
	if err := app.sessions.Current().SpawnFrom(msg, app.Conn); err != nil {
		return errors.Wrap(err, "adding client from osc message")
	}

	// Wait for announcement from the new client then respond to
	// the client who issued the add request.
	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case announcement := <-app.Announcements:
		m := "sending announcement response to client who requested the add operation"
		return errors.Wrap(app.SendTo(msg.Sender, announcement), m)
	}

	return nil
}
