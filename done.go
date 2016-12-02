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
	return errors.Wrap(app.SendTo(addr, osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(oscAddr),
			osc.String(nsm.DoneString),
		},
	}), "sending done message")
}
