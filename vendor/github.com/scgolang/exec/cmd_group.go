package exec

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// CmdGroup runs a set of commands.
type CmdGroup struct {
	addChan   chan *Cmd
	closeChan chan struct{}
	cmds      map[string]*Cmd
	ctx       context.Context
	group     *errgroup.Group
}

// NewCmdGroup creates a new CmdGroup instance.
func NewCmdGroup(ctx context.Context) *CmdGroup {
	g, gctx := errgroup.WithContext(ctx)

	cg := &CmdGroup{
		addChan:   make(chan *Cmd),
		closeChan: make(chan struct{}),
		cmds:      map[string]*Cmd{},
		ctx:       gctx,
		group:     g,
	}

	cg.Go(cg.Main)

	return cg
}

// Add adds a named Cmd to a CmdGroup.
func (cg *CmdGroup) Add(name string, cmd *Cmd) error {
	cmd.Name = name
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "getting stdout pipe")
	}
	cmd.stdoutPipe = stdout
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "starting command")
	}
	cg.Go(cmd.Wait)
	cg.addChan <- cmd
	return nil
}

// Close closes the CmdGroup.
func (cg *CmdGroup) Close() error {
	close(cg.closeChan)
	return nil
}

// Go runs a new goroutine as part of an errgroup.Group
func (cg *CmdGroup) Go(f func() error) {
	cg.group.Go(f)
}

// Main runs the background goroutine that manages the CmdGroup
func (cg *CmdGroup) Main() error {
ForLoop:
	for {
		select {
		case cmd := <-cg.addChan:
			cg.cmds[cmd.Name] = cmd
		case <-cg.closeChan:
			close(cg.addChan)
			break ForLoop
		}
	}
	return nil
}

// Wait waits for all commands to finish with an exit code of zero,
// or one to finish with non-zero exit code, whichever happens first.
func (cg *CmdGroup) Wait() error {
	return cg.group.Wait()
}
