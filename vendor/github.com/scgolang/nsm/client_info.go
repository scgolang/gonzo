package nsm

import (
	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// handleClientInfo runs a goroutine that handles the client to server informational messages.
// This is a persistent goroutine that closes the client's connection when it exits.
func (c *Client) handleClientInfo() error {
	for {
		select {
		case <-c.closedChan:
			return nil
		case <-c.ctx.Done():
			return c.ctx.Err()
		case isDirty := <-c.Session.Dirty():
			if err := c.sendDirty(isDirty); err != nil {
				return errors.Wrap(err, "send dirty message")
			}
		case isGUIShowing := <-c.Session.GUIShowing():
			if err := c.sendGUIShowing(isGUIShowing); err != nil {
				return errors.Wrap(err, "send gui-showing message")
			}
		case x := <-c.Session.Progress():
			if err := c.sendProgress(x); err != nil {
				return errors.Wrap(err, "send progress message")
			}
		case clientStatus := <-c.Session.ClientStatus():
			if err := c.sendClientStatus(clientStatus); err != nil {
				return errors.Wrap(err, "send client status message")
			}
		}
	}
}

// sendDirty sends an OSC message telling Non Session Manager
// if the client has unsaved changes.
func (c *Client) sendDirty(isDirty bool) error {
	var addr string
	if isDirty {
		addr = AddressClientIsDirty
	} else {
		addr = AddressClientIsClean
	}
	return c.Send(osc.Message{Address: addr}) // Will get wrapped by caller
}

// sendGUIShowing sends an OSC message telling Non Session Manager
// if the client has unsaved changes.
func (c *Client) sendGUIShowing(isGUIShowing bool) error {
	var addr string
	if isGUIShowing {
		addr = AddressClientGUIShowing
	} else {
		addr = AddressClientGUIHidden
	}
	return c.Send(osc.Message{Address: addr}) // Will get wrapped by caller
}

// sendProgress sends a progress measurement to Non Session Manager.
func (c *Client) sendProgress(x float32) error {
	return c.Send(osc.Message{
		Address: AddressClientProgress,
		Arguments: osc.Arguments{
			osc.Float(x),
		},
	})
}

// sendClientStatus sends a progress measurement to Non Session Manager.
func (c *Client) sendClientStatus(clientStatus ClientStatus) error {
	return c.Send(osc.Message{
		Address: AddressClientStatus,
		Arguments: osc.Arguments{
			osc.Int(int32(clientStatus.Priority)),
			osc.String(clientStatus.Message),
		},
	})
}
