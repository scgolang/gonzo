package main

import (
	"github.com/scgolang/osc"
)

// ClientLogs is an OSC method that returns the stdout of a client
// that is part of the current session.
func (app *App) ClientLogs(msg osc.Message) error {
	return nil
}
