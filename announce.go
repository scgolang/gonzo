package main

import (
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Announce handles the announcement of new clients.
func (app *App) Announce(msg osc.Message) error {
	app.debug("got announcement")

	// Add to client map.
	var (
		err = app.addClientFromAnnounce(msg)

		// default is successful response
		response = osc.Message{
			Address: nsm.AddressReply,
			Arguments: osc.Arguments{
				osc.String(nsm.AddressServerAnnounce),
				osc.String(ApplicationName),
				osc.String(app.Capabilities.String()),
			},
		}
	)

	// Respond with an error.
	if err != nil {
		var (
			addr   = nsm.AddressServerAnnounce
			code   = nsm.ErrGeneral
			errmsg = err.Error()
		)
		response = app.ReplyError(addr, code, errmsg)
	}

	// Send the response to the newly-announced client.
	err = app.SendTo(msg.Sender, response)

	// Send the announcement response on a channel.
	// This is how other clients who have requested the add operation
	// will find out about how the announcement handshake went
	select {
	case <-time.After(5 * time.Second):
		err = errors.Wrap(err, "timeout sending announcement response on a channel")
	case app.Announcements <- response:
	}

	return errors.Wrap(err, "sending response to new client")
}

// addClientFromAnnounce adds a client to the clients map from an announce message.
func (app *App) addClientFromAnnounce(msg osc.Message) error {
	if len(msg.Arguments) != 6 {
		return errors.New("expected 6 arguments in announce message")
	}
	appname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read appname in announce message")
	}
	capabilities, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read capabilities in announce message")
	}
	executableName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read executableName in announce message")
	}
	major, err := msg.Arguments[3].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read api major version in announce message")
	}
	minor, err := msg.Arguments[4].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read api minor version in announce message")
	}
	pid, err := msg.Arguments[5].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read pid in announce message")
	}
	key := Pid(pid)

	app.clientsMutex.RLock()
	if _, ok := app.clients[key]; ok {
		app.clientsMutex.RUnlock()
		return errors.Errorf("client with pid %d already exists", pid)
	}
	app.clientsMutex.RUnlock()

	app.clientsMutex.Lock()
	app.clients[key] = Client{
		ApplicationName: appname,
		Capabilities:    nsm.ParseCapabilities(capabilities),
		ExecutableName:  executableName,
		Major:           major,
		Minor:           minor,
	}
	app.clientsMutex.Unlock()

	app.debugf("added client to client map pid=%d name=%s major=%d minor=%d", pid, appname, major, minor)

	return nil
}
