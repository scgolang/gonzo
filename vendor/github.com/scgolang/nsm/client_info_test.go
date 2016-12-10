package nsm

import (
	"context"
	"testing"
	"time"
)

func TestClientDirty(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config    = testConfig()
		dirtyChan = make(chan bool)
	)
	config.Session = &mockSession{dirtyChan: dirtyChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case dirtyChan <- false:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case isDirty := <-nsmd.dirtyChan:
		if expected, got := isDirty, false; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}
}

func TestClientDirtyFailSend(t *testing.T) {
	var (
		// mockNsmd sets an environment variable to point the client to it's listening address
		nsmd      = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		config    = testConfig()
		dirtyChan = make(chan bool)
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config.Session = &mockSession{dirtyChan: dirtyChan}
	config.failSend = 2

	c := newClient(t, config)

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case dirtyChan <- true:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send dirty message: fail send`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientGuiShowing(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config         = testConfig()
		showingGuiChan = make(chan bool)
	)
	config.Session = &mockSession{showingGuiChan: showingGuiChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case showingGuiChan <- true:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case guiIsShowing := <-nsmd.guiShowingChan:
		if expected, got := guiIsShowing, true; expected != got {
			t.Fatalf("expected %t, got %t", expected, got)
		}
	}
}

func TestClientGuiShowingFailSend(t *testing.T) {
	var (
		// mockNsmd sets an environment variable to point the client to it's listening address
		nsmd           = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		config         = testConfig()
		showingGuiChan = make(chan bool)
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config.Session = &mockSession{showingGuiChan: showingGuiChan}
	config.failSend = 2

	c := newClient(t, config)

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case showingGuiChan <- false:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send gui-showing message: fail send`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientProgress(t *testing.T) {
	const val = float32(0.2323458)

	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config       = testConfig()
		progressChan = make(chan float32)
	)
	config.Session = &mockSession{progressChan: progressChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case progressChan <- val:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case x := <-nsmd.progressChan:
		if expected, got := val, x; expected != got {
			t.Fatalf("expected %f, got %f", expected, got)
		}
	}
}

func TestClientProgressFailSend(t *testing.T) {
	const val = float32(0.2323458)

	var (
		// mockNsmd sets an environment variable to point the client to it's listening address
		nsmd         = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		config       = testConfig()
		progressChan = make(chan float32)
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config.Session = &mockSession{progressChan: progressChan}
	config.failSend = 2

	c := newClient(t, config)

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case progressChan <- val:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send progress message: fail send`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestClientStatus(t *testing.T) {
	// mockNsmd sets an environment variable to point the client to it's listening address
	nsmd := newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
	defer func() { _ = nsmd.Close() }() // Best effort.

	var (
		config     = testConfig()
		statusChan = make(chan ClientStatus)
		val        = ClientStatus{Priority: PriorityHigh, Message: "bork"}
	)
	config.Session = &mockSession{statusChan: statusChan}

	if _, err := NewClient(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case statusChan <- val:
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case status := <-nsmd.statusChan:
		if expected, got := val, status; !expected.Equal(got) {
			t.Fatalf("expected %f, got %f", expected, got)
		}
	}
}

func TestClientStatusFailSend(t *testing.T) {
	var (
		// mockNsmd sets an environment variable to point the client to it's listening address
		nsmd       = newMockNsmd(t, mockNsmdConfig{listenAddr: "127.0.0.1:0"})
		config     = testConfig()
		statusChan = make(chan ClientStatus)
		val        = ClientStatus{Priority: PriorityHigh, Message: "bork"}
	)
	defer func() { _ = nsmd.Close() }() // Best effort.

	config.Session = &mockSession{statusChan: statusChan}
	config.failSend = 2

	c := newClient(t, config)

	// Setup a channel for the error from c.Wait
	errChan := make(chan error)
	go func() {
		errChan <- c.Wait()
	}()

	// Signal that the client has unsaved changes.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case statusChan <- val:
	}

	// Receive an error on the c.Wait channel.
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	case err := <-errChan:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if expected, got := `send client status message: fail send`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}
