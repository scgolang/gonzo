package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Sessions maintains a collection of sessions.
type Sessions struct {
	Home string // Home is the path to the directory that contains all the sessions.
	Dir  *os.File
	Curr string
	Mu   sync.RWMutex
	M    map[string]*Session

	ctx context.Context
	dbg Debugger
}

// NewSessions creates a new sessions collection.
func NewSessions(ctx context.Context, dbg Debugger, home string) (*Sessions, error) {
	s := &Sessions{
		Home: home,
		M:    map[string]*Session{},

		ctx: ctx,
		dbg: dbg,
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

// ListMessage creates an osc message that represents the sessions list.
func (s *Sessions) ListMessage() osc.Message {
	s.Mu.RLock()
	msg := osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(nsm.AddressServerSessions),
			osc.Int(len(s.M)),
		},
	}
	var (
		i            = 0
		curridx      = 0
		sessionNames = []osc.Argument{}
	)
	for name := range s.M {
		if name == s.Curr {
			curridx = i
		}
		sessionNames = append(sessionNames, osc.String(name))
		i++
	}
	s.Mu.RUnlock()
	msg.Arguments = append(msg.Arguments, osc.Int(curridx))
	msg.Arguments = append(msg.Arguments, sessionNames...)
	return msg
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
	sesh, err := NewSession(s.ctx, s.dbg, f)
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
	d, err := openOrCreateDir(s.Home)
	if err != nil {
		return errors.Wrap(err, "opening or creating "+s.Home)
	}
	s.Dir = d
	return nil
}

// Read reads sessions into memory.
func (s *Sessions) Read() error {
	if err := s.OpenHome(); err != nil {
		return errors.Wrap(err, "opening session home directory")
	}
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
		sesh, err := NewSession(s.ctx, s.dbg, f)
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

// Remove completely removes a session from disk.
func (s *Sessions) Remove(name string) error {
	var (
		exists      bool
		sesh        *Session
		sessionPath = filepath.Join(s.Home, name)
	)
	s.Mu.RLock()
	sesh, exists = s.M[sessionPath]
	s.Mu.RUnlock()

	if !exists {
		s.dbg.Debugf("sessions %#v\n", s.M)
		return errors.New("session " + sessionPath + " does not exist")
	}
	if sesh.Dirty() {
		return errors.New("session " + sessionPath + " has unsaved changes")
	}
	if err := os.RemoveAll(sessionPath); err != nil {
		return errors.Wrap(err, "removing "+sessionPath)
	}
	s.Mu.Lock()
	delete(s.M, sessionPath)
	s.Mu.Unlock()

	// If the removed session was the current session we need a new current session.
	if sessionPath == s.Curr {
		s.SelectCurrentRandomly()
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
	if len(s.M) == 0 {
		s.Curr = ""
	}
	for name := range s.M {
		s.Curr = name
		break
	}
	s.Mu.RUnlock()
	return nil
}
