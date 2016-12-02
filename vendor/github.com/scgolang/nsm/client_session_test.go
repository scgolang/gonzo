package nsm

import (
	"os"
	"testing"
	"time"
)

func TestClientSession(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	var (
		nsmd    = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		session = &mockSession{loadedChan: make(chan struct{})}
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	c, err := NewClient(ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      session,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	nsmd.SessionLoaded()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for session_is_loaded")
	case <-session.loadedChan:
	}
}
