package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const currentSessionCache = ".current"

// Session represents a session.
type Session struct {
	Path string
	Dir  *os.File
}

// NewSession creates a new session.
func NewSession(file string) (*Session, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrapf(err, "opening %s", file)
	}
	return &Session{Path: file, Dir: fd}, nil
}

// Open opens the session.
func (s *Session) Open() error {
	return nil
}

// Save saves the session.
func (s *Session) Save() error {
	return nil
}

// Sessions maintains a collection of sessions.
type Sessions struct {
	Home string // Home is the path to the directory that contains all the sessions.
	Dir  *os.File
	Curr string
	Mu   sync.RWMutex
	M    map[string]*Session
}

// NewSessions creates a new sessions collection.
func NewSessions(home string) (*Sessions, error) {
	s := &Sessions{
		Home: home,
		M:    map[string]*Session{},
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
	sesh, err := NewSession(f)
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
	for _, filename := range files {
		sesh, err := NewSession(filename)
		if err != nil {
			return errors.Wrapf(err, "reading %s", filename)
		}
		s.Mu.Lock()
		s.M[filename] = sesh
		s.Mu.Unlock()
	}

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
