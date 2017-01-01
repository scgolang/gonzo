package main

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/scgolang/exec"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

const currentSessionCache = ".current"

// Session represents a session.
type Session struct {
	clients      ClientMap
	clientsMutex sync.RWMutex

	cmdgrp *exec.Group

	Dir  *os.File
	Path string
}

// NewSession creates a new session.
func NewSession(file string, ctx context.Context) (*Session, error) {
	s := &Session{
		clients: ClientMap{},
		cmdgrp:  exec.NewGroup(ctx),
		Path:    file,
	}
	if err := s.initializeDirectory(); err != nil {
		return nil, errors.Wrap(err, "initializing session")
	}
	return s, nil
}

// Announce handles a client announcement.
func (s *Session) Announce(msg osc.Message) (*Client, error) {
	client, pid, err := s.clientFromAnnounce(msg)
	if err != nil {
		return nil, errors.Wrap(err, "creating client from announce message")
	}

	s.clientsMutex.RLock()
	if _, ok := s.clients[pid]; ok {
		s.clientsMutex.RUnlock()
		return nil, errors.Errorf("client with pid %d already exists", pid)
	}
	s.clientsMutex.RUnlock()

	s.clientsMutex.Lock()
	s.clients[pid] = client
	s.clientsMutex.Unlock()

	return client, nil
}

// Clients returns the session's ClientMap.
// It is very important to use this method as opposed to accessing
// the struct field directly since concurrent access to maps is not safe in Go.
func (s *Session) Clients() ClientMap {
	cm := ClientMap{}
	s.clientsMutex.RLock()
	for pid, c := range s.clients {
		cm[pid] = c
	}
	s.clientsMutex.RUnlock()
	return cm
}

// CreateCmdDirectory creates a directory for a process with the provided name.
func (s *Session) CreateCmdDirectory(cmdname string) error {
	cmdpath := filepath.Join(s.Path, cmdname)
	if _, err := openOrCreateDir(cmdpath); err != nil {
		return errors.Wrapf(err, "opening or creating %s", cmdpath)
	}
	return nil
}

// Dirty returns true if there are clients in the session with unsaved changes, false otherwise.
func (s *Session) Dirty() bool {
	// TODO
	return false
}

// Open opens the session.
func (s *Session) Open() error {
	return nil
}

// Goer can run goroutines based on func's that return an error.
type Goer interface {
	Go(func() error)
}

const (
	stdoutFilename = ".stdout"
	stderrFilename = ".stderr"
)

// PipeOutputFor pipes the output for the specified process to files in the session's directory.
func (s *Session) PipeOutputFor(cmdname string, g Goer) error {
	// Create the files where we will store the output.
	stdoutPath := filepath.Join(s.Path, cmdname, stdoutFilename)
	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return errors.Wrap(err, "creating "+stdoutPath)
	}
	stderrPath := filepath.Join(s.Path, cmdname, stderrFilename)
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return errors.Wrap(err, "creating "+stderrPath)
	}

	// Pipe the output to the newly created files.
	stdout, stderr, err := s.cmdgrp.Output(cmdname)
	if err != nil {
		return errors.Wrap(err, "getting output for "+cmdname)
	}
	g.Go(func() error {
		_, err := io.Copy(stdoutFile, stdout)
		return err
	})
	g.Go(func() error {
		_, err := io.Copy(stderrFile, stderr)
		return err
	})
	return nil
}

// Save saves the session.
func (s *Session) Save() error {
	return nil
}

// SpawnFrom spawns a new client based on an OSC message.
// We don't actually add the client to our client map until it announces itself successfully.
// This method returns the cmd name and nil or the empty string and an error.
func (s *Session) SpawnFrom(msg osc.Message, local net.Conn) (string, error) {
	if len(msg.Arguments) != 2 {
		return "", errors.New("add expects 2 arguments")
	}
	cmdname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return "", errors.Wrap(err, "could not read cmdname")
	}
	progname, err := msg.Arguments[1].ReadString()
	if err != nil {
		return "", errors.Wrap(err, "could not read progname")
	}
	var (
		cmd       = exec.Command(progname)
		localAddr = local.LocalAddr().String()
	)
	cmd.Env = append(os.Environ(), "NSM_URL="+localAddr)

	if err := s.cmdgrp.AddCmd(cmdname, cmd); err != nil {
		return "", errors.Wrap(err, "adding command "+progname)
	}
	return cmdname, nil
}

// clientFromAnnounce initializes a client from an announce message.
func (s *Session) clientFromAnnounce(msg osc.Message) (*Client, Pid, error) {
	if len(msg.Arguments) != 6 {
		return nil, 0, errors.New("expected 6 arguments in announce message")
	}

	client := &Client{}

	appname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read appname in announce message")
	}
	client.ApplicationName = appname

	capabilities, err := msg.Arguments[1].ReadString()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read capabilities in announce message")
	}
	client.Capabilities = nsm.ParseCapabilities(capabilities)

	executableName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read executableName in announce message")
	}
	client.ExecutableName = executableName

	major, err := msg.Arguments[3].ReadInt32()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read api major version in announce message")
	}
	client.Major = major

	minor, err := msg.Arguments[4].ReadInt32()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read api minor version in announce message")
	}
	client.Minor = minor

	pid, err := msg.Arguments[5].ReadInt32()
	if err != nil {
		return nil, 0, errors.Wrap(err, " read pid in announce message")
	}
	return client, Pid(pid), nil
}

// initializeDirectory initializes the session's directory.
// It attempts to open it and creates it if it doesn't exist.
func (s *Session) initializeDirectory() error {
	fd, err := openOrCreateDir(s.Path)
	if err != nil {
		return errors.Wrapf(err, "opening %s", s.Path)
	}
	s.Dir = fd
	return nil
}

const dirPerms = 0755

// openOrCreateDir opens a directory with the provided path,
// and creates it if it doesn't exist
func openOrCreateDir(dirpath string) (*os.File, error) {
	fd, err := os.Open(dirpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "opening %s", dirpath)
		}
		// Create the directory.
		if err := os.Mkdir(dirpath, dirPerms); err != nil {
			return nil, errors.Wrap(err, "making directory")
		}
		fd, err = os.Open(dirpath)
		if err != nil {
			return nil, errors.Wrap(err, "opening directory")
		}
	}
	return fd, nil
}
