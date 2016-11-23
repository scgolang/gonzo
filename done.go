package main

import (
	"net"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// done sends a reply using the provided osc.Conn to signal
// that we are done replying to addr.
func (app *App) done(addr net.Addr, oscAddr string) error {
	done, err := osc.NewMessage(nsm.AddressReply)
	if err != nil {
		return errors.Wrap(err, "create done reply")
	}
	if err := done.WriteString(oscAddr); err != nil {
		return errors.Wrap(err, "writing reply address")
	}
	if err := done.WriteString(nsm.DoneString); err != nil {
		return errors.Wrap(err, "writing done string")
	}
	if err := app.SendTo(addr, done); err != nil {
		errors.Wrap(err, "sending done message")
	}
	return nil
}
