package main

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// ListProjects replies with a list of projects.
func (app *App) ListProjects(msg *osc.Message) error {
	projects, err := app.readProjects()
	if err != nil {
		return errors.Wrap(err, "could not read projects")
	}

	addr := msg.Sender()
	sender, err := net.ResolveUDPAddr("udp", addr.String())
	conn, err := osc.DialUDP("udp", nil, sender)
	if err != nil {
		return errors.Wrapf(err, "could not connect to sender at %s", sender.String())
	}

	for _, project := range projects {
		reply, err := osc.NewMessage("/reply")
		if err != nil {
			return errors.Wrap(err, "could not create osc message")
		}
		if err := reply.WriteString(project); err != nil {
			return errors.Wrap(err, "could not add string to osc message")
		}
		if err := conn.Send(reply); err != nil {
			return errors.Wrap(err, "could not send reply")
		}
	}
	return nil
}

// readProjects reads the projects from the gonzo sessions directory.
func (app *App) readProjects() ([]string, error) {
	d, err := app.homeDir()
	if err != nil {
		return nil, errors.Wrap(err, "could not open home directory")
	}
	return d.Readdirnames(-1)
}

// homeDir tries to open the gonzo sessions directory.
// If it doesn't exist this method will create it.
func (app *App) homeDir() (*os.File, error) {
	d, err := os.Open(Home)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "could not open %s", Home)
		}
		if err := os.Mkdir(Home, 0755); err != nil {
			return nil, errors.Wrapf(err, "could not create %s", Home)
		}
		d, err = os.Open(Home)
		if err != nil {
			return nil, errors.Wrapf(err, "could not open %s", Home)
		}
	}
	return d, nil
}
