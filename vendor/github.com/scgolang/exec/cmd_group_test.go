package exec

import (
	"context"
	"io/ioutil"
	"testing"
)

func TestGroupAdd(t *testing.T) {
	group := NewGroup(context.Background())
	if err := group.Add("proc1", "echo", "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestGroupAddCmdFailStdoutAlreadySet(t *testing.T) {
	var (
		group = NewGroup(context.Background())
		echo  = Command("echo", "foo")
	)
	if _, err := echo.StdoutPipe(); err != nil {
		t.Fatal(err)
	}
	if err := group.AddCmd("echo", echo); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `getting stdout pipe: exec: Stdout already set`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestGroupAddCmdFailStderrAlreadySet(t *testing.T) {
	var (
		group = NewGroup(context.Background())
		echo  = Command("echo", "foo")
	)

	if _, err := echo.StderrPipe(); err != nil {
		t.Fatal(err)
	}
	if err := group.AddCmd("echo", echo); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `getting stderr pipe: exec: Stderr already set`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestGroupAddCmdFailLookupError(t *testing.T) {
	var (
		group = NewGroup(context.Background())
		echo  = Command("echolalialalialalia")
	)
	if err := group.AddCmd("echolalialalialalia", echo); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `starting command: exec: "echolalialalialalia": executable file not found in $PATH`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestGroupOutputMissingProcess(t *testing.T) {
	group := NewGroup(context.Background())
	_, _, err := group.Output("foo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := `process does not exist: foo`, err.Error(); expected != got {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestGroupProcessExists(t *testing.T) {
	var (
		group = NewGroup(context.Background())
		echo  = Command("echo", "foo")
	)
	if err := group.AddCmd("echo", echo); err != nil {
		t.Fatal(err)
	}
	if err := group.AddCmd("echo", echo); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		if expected, got := "command already exists: echo", err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}

func TestGroupOutput(t *testing.T) {
	var (
		group = NewGroup(context.Background())
		echo  = Command("echo", "foo")
	)
	if err := group.AddCmd("echo", echo); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := group.Output("echo")
	if err != nil {
		t.Fatal(err)
	}

	// Read all stdout.
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Fatal(err)
	}
	// Read all stderr.
	if _, err := ioutil.ReadAll(stderr); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err != nil {
		t.Fatal("expected error, got nil")
	}
	if expected, got := "foo\x0A", string(output); expected != got {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestGroupWait(t *testing.T) {
	var (
		group   = NewGroup(context.Background())
		nobueno = Command("false")
	)

	if err := group.AddCmd("false", nobueno); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err == nil {
		t.Fatal("expected error, got nil")
	}
}
