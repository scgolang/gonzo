package exec

import (
	"os/exec"
)

// ErrNotFound is the error resulting if a path search failed to find an executable file.
var ErrNotFound = exec.ErrNotFound

// LookPath searches for an executable binary named file in the directories named by the
// PATH environment variable.
// If file contains a slash, it is tried directly and the PATH is not consulted.
// The result may be an absolute path or a path relative to the current directory.
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Error records the name of a binary that failed to be executed and the reason it failed.
type Error struct {
	*exec.Error
}

// An ExitError reports an unsuccessful exit by a command.
type ExitError struct {
	*exec.ExitError
}
