package exec

import (
	"testing"
)

func TestLookPath(t *testing.T) {
	if _, err := LookPath("echolalialalialalia"); err == nil {
		t.Fatal("expected an error, got nil")
	} else {
		if expected, got := `exec: "echolalialalialalia": executable file not found in $PATH`, err.Error(); expected != got {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}
}
