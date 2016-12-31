package main

import (
	"net"

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
	if err := app.sessions.Read(); err != nil {
		return errors.Wrap(err, "read sessions")
	}

	app.debugf("read %d session(s)", len(app.sessions.M))

	msg := osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerSessions),
			osc.Int(len(app.sessions.M)),
		},
	}
	var (
		i            = 0
		curridx      = 0
		sessionNames = []osc.Argument{}
	)
	for name := range app.sessions.M {
		if name == app.sessions.Curr {
			curridx = i
		}
		sessionNames = append(sessionNames, osc.String(name))
		i++
	}
	msg.Arguments = append(msg.Arguments, osc.Int(curridx))
	msg.Arguments = append(msg.Arguments, sessionNames...)
	return errors.Wrapf(app.SendTo(addr, msg), "send %s reply", nsm.AddressServerSessions)
}
