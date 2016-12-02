package nsm

import (
	"os"
	"testing"

	"github.com/scgolang/osc"
)

func TestClientAnnounceReplyMissingArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	_, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      &mockSession{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `initialize client: announce app: handle announce reply: expected 4 arguments in announce reply, got 1`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
