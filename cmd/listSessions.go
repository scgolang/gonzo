package cmd

import (
	"net"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListSessions replies with a list of sessions.
func (app *App) ListSessions(msg osc.Message) error {
	app.Debug("listing sessions")

	// Read the sessions from disk and send each one as a reply message.
	if err := app.sendSessions(msg.Sender); err != nil {
		return errors.Wrap(err, "sending sessions")
	}
	return nil
}

// sendSessions sends the list of sessions as individual reply messages.
func (app *App) sendSessions(addr net.Addr) error {
	if err := app.sessions.Read(); err != nil {
		return errors.Wrap(err, "read sessions")
	}

	app.Debugf("read %d session(s)", len(app.sessions.M))

	msg := app.sessions.ListMessage()

	return errors.Wrapf(app.SendTo(addr, msg), "send %s reply", nsm.AddressServerSessions)
}
