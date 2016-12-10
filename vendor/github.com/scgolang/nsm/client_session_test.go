package nsm

import (
	"testing"
	"time"

	"github.com/scgolang/osc"
)

func TestClientSession(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	// This mock also sends an invalid announce reply (missing arguments)
	var (
		nsmd       = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		foobarChan = make(chan struct{})
		session    = &mockSession{
			loadedChan:  make(chan struct{}),
			showGuiChan: make(chan bool),
			methods: osc.Dispatcher{
				"/foo/bar": func(m osc.Message) error {
					close(foobarChan)
					return nil
				},
			},
		}
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config := testConfig()
	config.Session = session
	c := newClient(t, config)
	defer func() { _ = c.Close() }() // Best effort.

	// session_is_loaded message
	nsmd.SessionLoaded()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for session_is_loaded")
	case <-session.loadedChan:
	}

	// hide_optional_gui message
	nsmd.HideOptionalGUI()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for hide_optional_gui")
	case show := <-session.showGuiChan:
		if expected, got := false, show; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}

	// show_optional_gui message
	nsmd.ShowOptionalGUI()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for show_optional_gui")
	case show := <-session.showGuiChan:
		if expected, got := true, show; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}

	// test extra method added to client dispatcher
	if err := nsmd.SendTo(c.LocalAddr(), osc.Message{Address: "/foo/bar"}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for /foo/bar to be dispatched")
	case <-foobarChan:
	}
}
