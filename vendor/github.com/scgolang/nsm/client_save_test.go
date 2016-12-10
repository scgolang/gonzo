package nsm

import (
	"testing"

	"github.com/scgolang/osc"
)

func TestClientSave(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = &mockSession{
		save: mockReply{Message: "save successful"},
	}
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	reply := nsmd.SaveSession(osc.Message{
		Address: AddressClientSave,
		Arguments: osc.Arguments{
			osc.String("./test-projects"),
			osc.String("display_name"),
			osc.String("client_id"),
		},
	})
	if expected, got := 2, len(reply.Arguments); expected != got {
		t.Fatalf("expected %d arguments, got %d", expected, got)
	}
	replyAddr, err := reply.Arguments[0].ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := AddressClientSave, replyAddr; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	replyMessage, err := reply.Arguments[1].ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := `save successful`, replyMessage; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
