package main

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
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

	cmdgrp *exec.CmdGroup

	Dir  *os.File
	Path string
}

// NewSession creates a new session.
func NewSession(file string, ctx context.Context) (*Session, error) {
	s := &Session{
		clients: ClientMap{},
		cmdgrp:  exec.NewCmdGroup(ctx),
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

	if err := s.cmdgrp.Add(cmdname, cmd); err != nil {
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

// Sessions maintains a collection of sessions.
type Sessions struct {
	Home string // Home is the path to the directory that contains all the sessions.
	Dir  *os.File
	Curr string
	Mu   sync.RWMutex
	M    map[string]*Session

	ctx context.Context
}

// NewSessions creates a new sessions collection.
func NewSessions(home string, ctx context.Context) (*Sessions, error) {
	s := &Sessions{
		Home: home,
		M:    map[string]*Session{},

		ctx: ctx,
	}
	// Open the home dir.
	if err := s.OpenHome(); err != nil {
		return nil, errors.Wrap(err, "opening sessions home")
	}
	// Read the sessions.
	if err := s.Read(); err != nil {
		return nil, errors.Wrap(err, "reading sessions")
	}
	// Set the current session.
	if err := s.SelectCurrent(); err != nil {
		return nil, errors.Wrap(err, "select current session")
	}
	return s, nil
}

// Current returns the current session.
// There should always be a current session.
func (s *Sessions) Current() *Session {
	s.Mu.RLock()
	curr := s.M[s.Curr]
	s.Mu.RUnlock()
	return curr
}

// New creates a new session and makes it the current session.
func (s *Sessions) New(name string) error {
	f := filepath.Join(s.Home, name)

	// Return an error if the session already exists.
	s.Mu.RLock()
	if _, ok := s.M[f]; ok {
		s.Mu.RUnlock()
		return errors.Errorf("session already present %s", f)
	}
	s.Mu.RUnlock()

	// Create the new session and add it to the map.
	sesh, err := NewSession(f, s.ctx)
	if err != nil {
		return errors.Wrapf(err, "could not open session %s", f)
	}
	s.Mu.Lock()
	s.M[f] = sesh
	s.Mu.Unlock()

	return nil
}

// Close closes all the sessions.
func (s *Sessions) Close() error {
	// Write the current session.
	return nil
}

// OpenHome tries to open the sessions home directory, creating it if it doesn't exist.
func (s *Sessions) OpenHome() error {
	d, err := os.Open(s.Home)

	// Create the home directory if it doesn't exist.
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "open %s", s.Home)
		}
		if err := os.Mkdir(s.Home, 0755); err != nil {
			return errors.Wrapf(err, "create %s", s.Home)
		}
		d, err = os.Open(s.Home)
		if err != nil {
			return errors.Wrapf(err, "open %s", s.Home)
		}
	}
	s.Dir = d

	return nil
}

// Read reads sessions into memory.
func (s *Sessions) Read() error {
	// Read sessions and exit if there are none.
	files, err := s.Dir.Readdirnames(-1)
	if err != nil {
		return errors.Wrap(err, "reading directory contents")
	}
	if len(files) == 0 {
		return nil
	}

	m := map[string]*Session{}

	for _, filename := range files {
		f := filepath.Join(s.Home, filename)
		sesh, err := NewSession(f, s.ctx)
		if err != nil {
			return errors.Wrapf(err, "reading %s", filename)
		}
		m[f] = sesh
	}
	s.Mu.Lock()
	s.M = m
	s.Mu.Unlock()

	return nil
}

// SelectCurrent selects the current session based on either a file on disk
// or a random selection (if the file doesn't exist).
func (s *Sessions) SelectCurrent() error {
	f := filepath.Join(s.Home, currentSessionCache)
	fd, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			return s.SelectCurrentRandomly()
		}
		return errors.Wrapf(err, "could not open %s", f)
	}
	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		return errors.Wrapf(err, "reading %s", f)
	}
	curr := strings.TrimSpace(string(contents))

	s.Mu.RLock()
	if _, ok := s.M[curr]; !ok {
		s.Mu.RUnlock()
		return errors.Wrapf(err, "session %s does not exist", curr)
	}
	s.Curr = curr
	s.Mu.RUnlock()

	return nil
}

// SelectCurrentRandomly selects a session at random as the current session.
func (s *Sessions) SelectCurrentRandomly() error {
	s.Mu.RLock()
	for name := range s.M {
		s.Curr = name
		break
	}
	s.Mu.RUnlock()
	return nil
}
