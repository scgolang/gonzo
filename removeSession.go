package main

import (
	"fmt"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// RemoveSession removes a session.
// If the session being removed is the current session and has clients
// with unsaved changes then the session will not be removed an error reply will be sent.
func (app *App) RemoveSession(msg osc.Message) (string, nsm.Error) {
	const code = nsm.ErrUnsavedChanges

	if expected, got := 1, len(msg.Arguments); expected != got {
		return "", nsm.NewError(code, fmt.Sprintf("expected %d arguments, got %d", expected, got))
	}
	name, err := msg.Arguments[0].ReadString()
	if err != nil {
		return "", nsm.NewError(code, "reading string from message")
	}
	app.debugf("creating a new session named %s", name)

	if err := app.sessions.Remove(name); err != nil {
		return "", nsm.NewError(code, "removing session")
	}
	return "removed session " + name, nil
}
