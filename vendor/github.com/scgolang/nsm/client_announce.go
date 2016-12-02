package nsm

import (
	"os"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Announce announces a new nsm application.
func (c *Client) Announce() error {
	// Send the announce message.
	msg, err := c.newAnnounceMsg()
	if err != nil {
		return errors.Wrap(err, "create announce message")
	}
	if err := c.Send(msg); err != nil {
		return errors.Wrap(err, "send announce message")
	}

	// Wait for the server's reply.
	reply, err := c.wait(AddressServerAnnounce)
	if err != nil {
		return errors.Wrap(err, "waiting for announce reply")
	}
	return errors.Wrap(c.handleAnnounce(reply), "handle announce reply")
}

// newAnnounceMsg creates a new announce message.
func (c *Client) newAnnounceMsg() (osc.Message, error) {
	msg := osc.Message{
		Address: AddressServerAnnounce,
		Arguments: osc.Arguments{
			osc.String(os.Args[0]),
			osc.String(c.Capabilities.String()),
			osc.String(os.Args[0]),
			osc.Int(c.Major),
			osc.Int(c.Minor),
			osc.Int(int32(c.PID)),
		},
	}
	// Write name.
	if c.Name != "" {
		msg.Arguments[0] = osc.String(c.Name)
	}
	return msg, nil
}

// handleAnnounce handles a reply to the announce message.
func (c *Client) handleAnnounce(msg osc.Message) error {
	if got := len(msg.Arguments); got != 4 {
		return errors.Errorf("expected 4 arguments in announce reply, got %d", got)
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "read reply address")
	}
	if addr != AddressServerAnnounce {
		// TODO: put the message back in a queue and keep waiting
		return nil
	}

	serverMsg, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "read reply message")
	}
	smName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "read session manager name")
	}
	capsRaw, err := msg.Arguments[3].ReadString()
	if err != nil {
		return errors.Wrap(err, "read session manager capabilities")
	}
	return c.Session.Announce(ServerInfo{
		Message:      serverMsg,
		ServerName:   smName,
		Capabilities: ParseCapabilities(capsRaw),
	})
}
