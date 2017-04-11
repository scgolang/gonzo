package cmd

import (
	"net"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListClients replies with a list of clients.
func (app *App) ListClients(msg osc.Message) error {
	app.Debug("listing clients")

	// Read the clients from disk and send each one as a reply message.
	if err := app.sendClients(msg.Sender); err != nil {
		return errors.Wrap(err, "sending clients")
	}
	return nil
}

// sendClients sends the list of clients as individual reply messages.
func (app *App) sendClients(addr net.Addr) error {
	clients := app.sessions.Current().Clients()

	msg := osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerClients),
			osc.Int(len(clients)),
		},
	}

	for pid, client := range clients {
		msg.Arguments = append(msg.Arguments, []osc.Argument{
			osc.String(client.ApplicationName),
			osc.String(client.Capabilities.String()),
			osc.String(client.ExecutableName),
			osc.Int(client.Major),
			osc.Int(client.Minor),
			osc.Int(pid),
		}...)
		app.Debugf("added client to message pid=%d name=%s", pid, client.ApplicationName)
	}
	return errors.Wrapf(app.SendTo(addr, msg), "send %s reply", nsm.AddressServerClients)
}
