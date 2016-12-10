package main

import (
	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Announce announces new clients.
func (app *App) Announce(msg osc.Message) error {
	app.debug("got announcement")

	// Add to client map.
	err := app.addClientFromAnnounce(msg)
	if replyErr := app.replyAnnounce(msg, err); replyErr != nil {
	}
	return err
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
	app.clients[Pid(pid)] = Client{
		ApplicationName: appname,
		Capabilities:    nsm.ParseCapabilities(capabilities),
		ExecutableName:  executableName,
		Major:           major,
		Minor:           minor,
	}
	return nil
}

// replyAnnounce replies to an announce message.
func (app *App) replyAnnounce(msg osc.Message, err error) error {
	if err == nil {
		return errors.Wrap(app.SendTo(msg.Sender, osc.Message{
			Address: nsm.AddressReply,
			Arguments: osc.Arguments{
				osc.String(nsm.AddressServerAnnounce),
			},
		}), "")
	}
	return errors.Wrap(app.SendTo(msg.Sender, osc.Message{
		Address: nsm.AddressError,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerAnnounce),
		},
	}), "")
}
