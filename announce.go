package main

import (
	"time"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Announce handles the announcement of new clients.
func (app *App) Announce(msg osc.Message) nsm.Error {
	app.debug("got announcement")

	// Add to client map.
	if err := app.sessions.Current().Announce(msg); err != nil {
		return nsm.NewError(nsm.ErrLaunchFailed, err.Error())
	}

	// default is successful response
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
		return nsm.NewError(nsm.ErrGeneral, err.Error())
	}

	// Send the announcement response on a channel.
	// This is how other clients who have requested the add operation
	// will find out about how the announcement handshake went
	select {
	case <-time.After(5 * time.Second):
		return nsm.NewError(nsm.ErrGeneral, "timeout sending announcement response on a channel")
	case app.Announcements <- response:
	}
	return nil
}
