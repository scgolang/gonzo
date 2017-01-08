package main

import (
	"fmt"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

const (
	stdoutArg = int32(1)
	stderrArg = int32(2)
)

// ClientLogs is an OSC method that returns the stdout or stderr of a client that is part of the current session.
func (app *App) ClientLogs(msg osc.Message) (string, nsm.Error) {
	const code = nsm.ErrGeneral

	if expected, got := 2, len(msg.Arguments); expected != got {
		return "", nsm.NewError(code, fmt.Sprintf("expected %d arguments, got %d", expected, got))
	}
	clientName, err := msg.Arguments[0].ReadString()
	if err != nil {
		return "", nsm.NewError(code, fmt.Sprintf("reading clientName: %s", err))
	}
	fd, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return "", nsm.NewError(code, fmt.Sprintf("reading file_descriptor argument: %s", err))
	}
	if fd != stdoutArg && fd != stderrArg {
		return "", nsm.NewError(code, fmt.Sprintf("file_descriptor argument must be either %d or %d", stderrArg, stdoutArg))
	}

	app.Debugf("getting logs for %s", clientName)

	currentSession := app.sessions.Current()
	reply, err := currentSession.Logs(clientName, fd)
	if err != nil {
		return "", nsm.NewError(code, fmt.Sprintf("getting client logs: %s", err))
	}
	if err := app.SendTo(msg.Sender, reply); err != nil {
		return "", nsm.NewError(code, fmt.Sprintf("sending reply: %s", err))
	}
	return "sent logs for client " + clientName, nil
}
