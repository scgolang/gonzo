package main

import (
	"fmt"
	"time"

	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Announce handles the announcement of new clients.
func (app *App) Announce(msg osc.Message) nsm.Error {
	app.debug("got announcement")

	// Add to client map.
	if err := app.addClientFromAnnounce(msg); err != nil {
		return err
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

// addClientFromAnnounce adds a client to the clients map from an announce message.
func (app *App) addClientFromAnnounce(msg osc.Message) nsm.Error {
	const code = nsm.ErrLaunchFailed

	if len(msg.Arguments) != 6 {
		return nsm.NewError(code, "expected 6 arguments in announce message")
	}
	appname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return nsm.NewError(code, "could not read appname in announce message")
	}
	capabilities, err := msg.Arguments[1].ReadString()
	if err != nil {
		return nsm.NewError(code, "could not read capabilities in announce message")
	}
	executableName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return nsm.NewError(code, "could not read executableName in announce message")
	}
	major, err := msg.Arguments[3].ReadInt32()
	if err != nil {
		return nsm.NewError(code, "could not read api major version in announce message")
	}
	minor, err := msg.Arguments[4].ReadInt32()
	if err != nil {
		return nsm.NewError(code, "could not read api minor version in announce message")
	}
	pid, err := msg.Arguments[5].ReadInt32()
	if err != nil {
		return nsm.NewError(code, "could not read pid in announce message")
	}
	key := Pid(pid)

	app.clientsMutex.RLock()
	if _, ok := app.clients[key]; ok {
		app.clientsMutex.RUnlock()
		return nsm.NewError(code, fmt.Sprintf("client with pid %d already exists", pid))
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
