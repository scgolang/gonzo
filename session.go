package main

import (
	"context"
	"net"
	"os"
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
	if err := s.initialize(); err != nil {
		return nil, errors.Wrap(err, "initializing session")
	}
	return s, nil
}

// Announce handles a client announcement.
func (s *Session) Announce(msg osc.Message) error {
	if len(msg.Arguments) != 6 {
		return errors.New("expected 6 arguments in announce message")
	}
	appname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read appname in announce message")
	}
	capabilities, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read capabilities in announce message")
	}
	executableName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read executableName in announce message")
	}
	major, err := msg.Arguments[3].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read api major version in announce message")
	}
	minor, err := msg.Arguments[4].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read api minor version in announce message")
	}
	pid, err := msg.Arguments[5].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read pid in announce message")
	}
	key := Pid(pid)

	s.clientsMutex.RLock()
	if _, ok := s.clients[key]; ok {
		s.clientsMutex.RUnlock()
		return errors.Errorf("client with pid %d already exists", pid)
	}
	s.clientsMutex.RUnlock()

	s.clientsMutex.Lock()
	s.clients[key] = Client{
		ApplicationName: appname,
		Capabilities:    nsm.ParseCapabilities(capabilities),
		ExecutableName:  executableName,
		Major:           major,
		Minor:           minor,
	}
	s.clientsMutex.Unlock()

	return nil
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

// Open opens the session.
func (s *Session) Open() error {
	return nil
}

// Save saves the session.
func (s *Session) Save() error {
	return nil
}

// SpawnFrom spawns a new client based on an OSC message.
// We don't actually add the client to our client map until it
// announces itself successfully.
func (s *Session) SpawnFrom(msg osc.Message, local net.Conn) error {
	if len(msg.Arguments) != 2 {
		return errors.New("add expects 2 arguments")
	}
	cmdname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read cmdname")
	}
	progname, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read progname")
	}

	var (
		cmd       = exec.Command(progname)
		localAddr = local.LocalAddr().String()
	)
	cmd.Env = append(os.Environ(), "NSM_URL="+localAddr)

	if err := s.cmdgrp.AddCmd(cmdname, cmd); err != nil {
		return errors.Wrap(err, "adding command "+progname)
	}

	return nil
}

// initialize initializes the session.
func (s *Session) initialize() error {
	fd, err := os.Open(s.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "opening %s", s.Path)
		}
		// Create the directory.
		if err := os.Mkdir(s.Path, 0755); err != nil {
			return errors.Wrap(err, "making directory")
		}
		fd, err = os.Open(s.Path)
		if err != nil {
			return errors.Wrap(err, "opening directory")
		}
	}
	s.Dir = fd
	return nil
}
