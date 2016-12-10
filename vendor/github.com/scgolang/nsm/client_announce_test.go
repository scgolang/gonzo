package nsm

import (
	"context"
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

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: expected 4 arguments in announce reply, got 1`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceReplyWrongAddress(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String("/foo/bar"),
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: expected /nsm/server/announce, got /foo/bar`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceReplyFirstArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: read reply first argument: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceReplySecondArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
				osc.Int(3),
				osc.Int(3),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: read reply message: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceReplyThirdArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
				osc.String("This reply message is false."),
				osc.Int(3),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: read session manager name: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceReplyFourthArgumentWrongType(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
				osc.String("This reply message is false."),
				osc.String("Mr. Session Manager"),
				osc.Int(3),
			},
		},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: read session manager capabilities: invalid type tag`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientAnnounceSendError(t *testing.T) {
	var (
		// mockNsmd sets an environment variable to point the client to it's listening address
		nsmd   = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		config = testConfig()
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config.failSend = 1

	_, err := NewClient(context.Background(), config)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `initialize client: announce app: send announce message: fail send`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
