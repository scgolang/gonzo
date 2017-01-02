package main

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

const (
	stdoutArg = int32(1)
	stderrArg = int32(2)
)

// ClientLogs is an OSC method that returns the stdout or stderr of a client
// that is part of the current session.
func (app *App) ClientLogs(msg osc.Message) error {
	if expected, got := 2, len(msg.Arguments); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	clientName, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading clientName")
	}
	fd, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading file_descriptor")
	}
	if fd != stdoutArg && fd != stderrArg {
		return errors.Wrapf(err, "file_descriptor must be either %d or %d", stdoutArg, stderrArg)
	}
	app.debugf("getting logs for %s", clientName)
	return nil
}
