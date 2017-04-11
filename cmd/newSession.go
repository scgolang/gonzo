package cmd

import (
	"fmt"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// NewSession creates a new session, and makes the new session the current session.
func (app *App) NewSession(msg osc.Message) (string, nsm.Error) {
	const code = nsm.ErrNoSessionOpen

	if expected, got := 1, len(msg.Arguments); expected != got {
		return "", nsm.NewError(code, fmt.Sprintf("expected %d arguments, got %d", expected, got))
	}
	name, err := msg.Arguments[0].ReadString()
	if err != nil {
		return "", nsm.NewError(code, "reading string from message")
	}
	app.Debugf("creating a new session named %s", name)

	if err := app.sessions.New(name); err != nil {
		return "", nsm.NewError(code, "creating new session")
	}
	return "created new session " + name, nil
}
