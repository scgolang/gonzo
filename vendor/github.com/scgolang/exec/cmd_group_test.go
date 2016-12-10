package exec

import (
	"context"
	"testing"
)

func TestCmdGroupAddCmd(t *testing.T) {
	var (
		group = NewCmdGroup(context.Background())
		echo  = Command("echo", "foo")
	)
	defer group.Close()

	if err := group.Add("echo", echo); err != nil {
		t.Fatal(err)
	}
}

func TestCmdGroupAddCmdFailStdoutAlreadySet(t *testing.T) {
	var (
		group = NewCmdGroup(context.Background())
		echo  = Command("echo", "foo")
	)
	defer group.Close()

	if _, err := echo.StdoutPipe(); err != nil {
		t.Fatal(err)
	}
	if err := group.Add("echo", echo); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `getting stdout pipe: exec: Stdout already set`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestCmdGroupAddCmdFailLookupError(t *testing.T) {
	var (
		group = NewCmdGroup(context.Background())
		echo  = Command("echolalialalialalia")
	)
	defer group.Close()

	if err := group.Add("echolalialalialalia", echo); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `starting command: exec: "echolalialalialalia": executable file not found in $PATH`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestCmdGroupWait(t *testing.T) {
	var (
		group   = NewCmdGroup(context.Background())
		nobueno = Command("false")
	)

	if err := group.Add("false", nobueno); err != nil {
		t.Fatal(err)
	}
	if err := group.Close(); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err == nil {
		t.Fatal("expected error, got nil")
	}
}
