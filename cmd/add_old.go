package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Add starts a new client program.
func (app *App) Add(msg osc.Message) error {
	currentSession := app.sessions.Current()
	cmdname, err := currentSession.SpawnFrom(msg, app.Conn)
	if err != nil {
		return errors.Wrap(err, "adding client from osc message")
	}
	if err := currentSession.CreateCmdDirectory(cmdname); err != nil {
		return errors.Wrap(err, "creating directory for new client")
	}
	if err := currentSession.PipeOutputFor(cmdname, app.errgrp); err != nil {
		return errors.Wrap(err, "piping output from new client")
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
