package main

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListProjects replies with a list of projects.
func (app *App) ListProjects(msg *osc.Message) error {
	// Setup a connection to the sender.
	addr := msg.Sender()
	sender, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		return errors.Wrapf(err, "resolve sender address from %s", addr.String())
	}
	conn, err := osc.DialUDP("udp", nil, sender)
	if err != nil {
		return errors.Wrapf(err, "connect to sender at %s", sender.String())
	}

	// Read the projects from disk and send each one as a reply message.
	if err := app.sendProjects(conn); err != nil {
		return errors.Wrap(err, "sending projects")
	}

	// Signal the client we are done.
	done, err := osc.NewMessage(nsm.AddressReply)
	if err != nil {
		return errors.Wrap(err, "create done reply")
	}
	if err := done.WriteString(nsm.AddressServerList); err != nil {
		return errors.Wrap(err, "writing reply address")
	}
	return errors.Wrap(conn.Send(done), "sending done message")
}

// sendProjects sends the list of projects as individual reply messages.
func (app *App) sendProjects(conn osc.Conn) error {
	projects, err := app.readProjects()
	if err != nil {
		return errors.Wrap(err, "read projects")
	}
	for _, project := range projects {
		reply, err := osc.NewMessage(nsm.AddressReply)
		if err != nil {
			return errors.Wrap(err, "create osc message")
		}
		if err := reply.WriteString(nsm.AddressServerList); err != nil {
			return errors.Wrap(err, "writing reply address")
		}
		if err := reply.WriteString(project); err != nil {
			return errors.Wrap(err, "add string to osc message")
		}
		if err := conn.Send(reply); err != nil {
			return errors.Wrap(err, "send reply")
		}
	}
	return nil
}

// readProjects reads the projects from the gonzo sessions directory.
func (app *App) readProjects() ([]string, error) {
	d, err := app.homeDir()
	if err != nil {
		return nil, errors.Wrap(err, "open home directory")
	}
	return d.Readdirnames(-1)
}

// homeDir tries to open the gonzo sessions directory.
// If it doesn't exist this method will create it.
func (app *App) homeDir() (*os.File, error) {
	d, err := os.Open(Home)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "open %s", Home)
		}
		if err := os.Mkdir(Home, 0755); err != nil {
			return nil, errors.Wrapf(err, "create %s", Home)
		}
		d, err = os.Open(Home)
		if err != nil {
			return nil, errors.Wrapf(err, "open %s", Home)
		}
	}
	return d, nil
}
