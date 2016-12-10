package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/exec"
	"github.com/scgolang/osc"
)

// Add starts a new client program.
func (app *App) Add(msg osc.Message) error {
	if len(msg.Arguments) != 2 {
		return errors.New("add expects 2 arguments")
	}
	cmdname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read cmdname")
	}
	progname, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read progname")
	}

	var (
		cmd       = exec.Command(progname)
		localAddr = app.Conn.LocalAddr().String()
	)
	cmd.Env = append(os.Environ(), "NSM_URL="+localAddr)

	app.debugf("adding %s", cmdname)
	if err := app.cmdgrp.Add(cmdname, cmd); err != nil {
		return errors.Wrap(err, "adding command "+progname)
	}
	app.debugf("added %s", cmdname)

	// HACK: osc server needs to support non-blocking method dispatch
	app.Go(func() error {
		// Wait for announcement from the new client then respond to
		// the client who issued the add request.
		select {
		case <-time.After(2 * time.Second):
			return errors.New("timeout")
		case announcement := <-app.Announcements:
			m := "sending announcement response to client who requested the add operation"
			return errors.Wrap(app.SendTo(msg.Sender, announcement), m)
		}
	})

	return nil
}
