package main

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// NewSession creates a new session, and makes the new session the current session.
func (app *App) NewSession(msg osc.Message) error {
	if expected, got := 1, len(msg.Arguments); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	name, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading string from message")
	}
	app.debugf("creating a new session named %s", name)

	return errors.Wrap(app.sessions.New(name), "creating new session")
}
