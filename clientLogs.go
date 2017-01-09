package main

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

const (
	stdoutArg = int32(1)
	stderrArg = int32(2)
)

// ClientLogs is an OSC method that returns the stdout or stderr of a client that is part of the current session.
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
		return errors.Wrap(err, "reading file_descriptor argument")
	}
	if fd != stdoutArg && fd != stderrArg {
		return errors.Errorf("file_descriptor argument must be either %d or %d", stderrArg, stdoutArg)
	}

	app.Debugf("getting logs for %s", clientName)

	currentSession := app.sessions.Current()
	reply, err := currentSession.Logs(clientName, fd)
	if err != nil {
		return errors.Wrap(err, "getting client logs")
	}
	app.Debugf("sending client logs reply %#v", reply)
	app.Debugf("sending reply to %s", msg.Sender.String())

	if err := app.SendTo(msg.Sender, reply); err != nil {
		app.Debugf("error sending reply %s", err)
		return errors.Wrap(err, "sending reply")
	}
	app.Debugf("sent reply to %s", msg.Sender.String())

	return nil
}
