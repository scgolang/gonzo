package main

import (
	"os/exec"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Add starts a new client program.
func (app *App) Add(msg *osc.Message) error {
	progname, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read progname")
	}
	var (
		cmd       = exec.Command(progname)
		localAddr = app.Conn.LocalAddr().String()
	)
	cmd.Env = []string{
		"NSM_URL=" + localAddr,
	}
	app.Go(cmd.Run)
	return nil
}