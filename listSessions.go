package main

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListSessions replies with a list of sessions.
func (app *App) ListSessions(msg osc.Message) error {
	app.debug("listing sessions")

	// Read the sessions from disk and send each one as a reply message.
	if err := app.sendSessions(msg.Sender); err != nil {
		return errors.Wrap(err, "sending sessions")
	}
	return nil
}

// sendSessions sends the list of sessions as individual reply messages.
func (app *App) sendSessions(addr net.Addr) error {
	sessions, err := app.readSessions()
	if err != nil {
		return errors.Wrap(err, "read sessions")
	}

	app.debugf("read %d session(s)", len(sessions))

	msg := osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerSessions),
			osc.Int(len(sessions)),
		},
	}
	for _, session := range sessions {
		msg.Arguments = append(msg.Arguments, osc.String(session))
	}
	return errors.Wrapf(app.SendTo(addr, msg), "send %s reply", nsm.AddressServerSessions)
}

// readSessions reads the sessions from the gonzo sessions directory.
func (app *App) readSessions() ([]string, error) {
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
