package nsm

import (
	"testing"
)

func TestSessionInfo(t *testing.T) {
	si := SessionInfo{}

	if err := si.Announce(ServerInfo{}); err != nil {
		t.Fatal(err)
	}
	if err := si.IsLoaded(); err != nil {
		t.Fatal(err)
	}
	if err := si.ShowGUI(false); err != nil {
		t.Fatal(err)
	}
	if ch := si.Dirty(); ch != nil {
		t.Fatalf("expected nil channel, got %+v", ch)
	}
	if ch := si.GUIShowing(); ch != nil {
		t.Fatalf("expected nil channel, got %+v", ch)
	}
	if ch := si.Progress(); ch != nil {
		t.Fatalf("expected nil channel, got %+v", ch)
	}
	if ch := si.ClientStatus(); ch != nil {
		t.Fatalf("expected nil channel, got %+v", ch)
	}
	if expected, got := 0, len(si.Methods()); expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
}
