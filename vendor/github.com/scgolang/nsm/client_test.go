package nsm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/scgolang/osc"
)

func testConfig() ClientConfig {
	return ClientConfig{
		Name:                 "test_client",
		Capabilities:         Capabilities{"switch", "progress"},
		Major:                1,
		Minor:                2,
		PID:                  os.Getpid(),
		Session:              &mockSession{},
		Timeout:              DefaultTimeout,
		WaitForAnnounceReply: true,
	}
}

func newClient(t *testing.T, config ClientConfig) *Client {
	c, err := NewClient(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClientErrNilSession(t *testing.T) {
	if _, err := NewClient(context.Background(), ClientConfig{}); err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %+v", err)
	}
}

func TestClientAnnounce(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if err := os.Unsetenv(NsmURL); err != nil {
		t.Fatalf("unset NSM_URL %s", err)
	}

	c := newClient(t, ClientConfig{
		Name:         "test_client",
		Capabilities: Capabilities{"switch", "progress"},
		Major:        1,
		Minor:        2,
		PID:          os.Getpid(),
		Session:      &mockSession{},
		NsmURL:       nsmd.LocalAddr().String(),
	})
	defer func() { _ = c.Close() }() // Best effort.
}

func TestClientAnnounceTimeout(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announcePause: 100 * time.Millisecond,
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	_, err := NewClient(context.Background(), ClientConfig{
		Name:                 "test_client",
		Capabilities:         Capabilities{"switch", "progress"},
		Major:                1,
		Minor:                2,
		PID:                  os.Getpid(),
		Session:              &mockSession{},
		Timeout:              10 * time.Millisecond,
		WaitForAnnounceReply: true,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `initialize client: announce app: timeout`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestClientNoNsmUrl(t *testing.T) {
	if err := os.Unsetenv(NsmURL); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: No `+NsmURL+` environment variable`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGarbageNsmUrl(t *testing.T) {
	if err := os.Setenv(NsmURL, "garbage"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: resolve udp remote address: missing port in address garbage`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGarbageListenAddr(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announcePause: 100 * time.Millisecond,
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.ListenAddr = "garbage"
	if _, err := NewClient(context.Background(), config); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: dial udp: resolve udp listening address: missing port in address garbage`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientReplyNoArguments(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr:    "127.0.0.1:0",
		announceReply: osc.Message{Address: AddressReply},
	})
	defer func() { _ = nsmd.Close() }() // Best effort.

	if _, err := NewClient(context.Background(), testConfig()); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := `initialize client: announce app: handle announce reply: expected 4 arguments in announce reply, got 0`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientReplyFirstArgumentWrongAddress(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{
		listenAddr: "127.0.0.1:0",
		announceReply: osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String("/foo/bar"),
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

func TestClientContextTimeout(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	c, err := NewClient(ctx, testConfig())
	if err != nil {
		t.Fatal("expected error, got nil")
	}
	defer func() { _ = c.Close() }() // Best effort.
	defer cancel()

	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err != context.DeadlineExceeded {
			t.Fatalf("expected context.DeadlineExceeded, got %+v", err)
		}
	}
}
