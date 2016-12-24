package exec

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Group runs a set of commands.
type Group struct {
	cmds      map[string]*Cmd
	cmdsMutex sync.RWMutex
	ctx       context.Context
}

// NewGroup creates a new Group instance.
func NewGroup(ctx context.Context) *Group {
	return &Group{
		cmds: map[string]*Cmd{},
		ctx:  ctx,
	}
}

// Add adds a named Cmd to a Group.
func (cg *Group) Add(name, command string, args ...string) error {
	return cg.AddCmd(name, CommandContext(cg.ctx, command, args...))
}

// AddCmd adds the provided command to the group.
func (cg *Group) AddCmd(name string, cmd *Cmd) error {
	cg.cmdsMutex.Lock()
	if _, ok := cg.cmds[name]; ok {
		cg.cmdsMutex.Unlock()
		return errors.New("command already exists: " + name)
	}
	cg.cmds[name] = nil
	cg.cmdsMutex.Unlock()

	cmd.Name = name

	// Get the stderr pipe.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "getting stderr pipe")
	}
	cmd.stderrPipe = stderr

	// Get the stdout pipe.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "getting stdout pipe")
	}
	cmd.stdoutPipe = stdout

	// Start the process and wait for it in a new goroutine.
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "starting command")
	}

	// Add it to the map.
	cg.cmdsMutex.Lock()
	cg.cmds[name] = cmd
	cg.cmdsMutex.Unlock()

	return nil
}

// Output returns the stdout and stderr pipes for the named process.
func (cg *Group) Output(name string) (stdout, stderr io.ReadCloser, err error) {
	var (
		cmd    *Cmd
		exists bool
	)
	cg.cmdsMutex.RLock()
	cmd, exists = cg.cmds[name]
	cg.cmdsMutex.RUnlock()

	if !exists {
		return nil, nil, errors.New("process does not exist: " + name)
	}
	return cmd.stdoutPipe, cmd.stderrPipe, nil
}

// Wait waits for all commands to finish.
func (cg *Group) Wait() error {
	cmds := []*Cmd{}
	cg.cmdsMutex.RLock()
	for _, cmd := range cg.cmds {
		cmds = append(cmds, cmd)
	}
	cg.cmdsMutex.RUnlock()

	errs := []string{}
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, " and "))
}
