package nsm

import (
	"os"
	"testing"

	"github.com/scgolang/osc"
)

func TestClientOpenReplyMissingArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	c, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session: &mockSession{
			open: mockReply{Message: "session started"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	reply := nsmd.OpenSession(osc.Message{
		Address: AddressClientOpen,
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
	if expected, got := AddressClientOpen, replyAddr; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	replyMessage, err := reply.Arguments[1].ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := `session started`, replyMessage; expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
