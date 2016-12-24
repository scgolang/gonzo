package exec

import (
	"context"
	"io"
	"os/exec"
)

// Cmd represents an external command being prepared or run.
//
// A Cmd cannot be reused after calling its Run, Output or CombinedOutput methods.
//
type Cmd struct {
	*exec.Cmd

	Name       string
	stderrPipe io.ReadCloser
	stdoutPipe io.ReadCloser
}

// Command returns the Cmd struct to execute the named program with the given arguments.
//
// It sets only the Path and Args in the returned structure.
//
// If name contains no path separators, Command uses LookPath to resolve the path to
// a complete name if possible. Otherwise it uses name directly.
//
// The returned Cmd's Args field is constructed from the command name followed by
// the elements of arg, so arg should not include the command name itself.
// For example, Command("echo", "hello")
//
func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd: exec.Command(name, arg...),
	}
}

// CommandContext is like Command but includes a context.
//
// The provided context is used to kill the process (by calling os.Process.Kill)
// if the context becomes done before the command completes on its own.
//
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	return &Cmd{
		Cmd: exec.CommandContext(ctx, name, arg...),
	}
}
