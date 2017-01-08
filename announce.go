package main

import (
	"time"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Announce handles the announcement of new clients.
func (app *App) Announce(msg osc.Message) (string, nsm.Error) {
	app.Debug("got announcement")

	// Add to client map.
	client, err := app.sessions.Current().Announce(msg)
	if err != nil {
		return "", nsm.NewError(nsm.ErrLaunchFailed, err.Error())
	}

	// Default is successful response.
	response := osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerAnnounce),
			osc.String(ApplicationName),
			osc.String(app.Capabilities.String()),
		},
	}

	// Send the response to the newly-announced client.
	if err := app.SendTo(msg.Sender, response); err != nil {
		return "", nsm.NewError(nsm.ErrGeneral, err.Error())
	}

	// Send the announcement response on a channel.
	// This is how other clients who have requested the add operation
	// will find out about how the announcement handshake went
	select {
	case <-time.After(5 * time.Second):
		return "", nsm.NewError(nsm.ErrGeneral, "timeout sending announcement response on a channel")
	case app.Announcements <- response:
	}
	return "successful announcement from " + client.ApplicationName, nil
}
