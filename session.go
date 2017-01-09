package main

import (
	"bufio"
	"context"
	"io"
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

type sessionClient struct {
	stderrPath string
	stdoutPath string
}

// Session represents a session.
type Session struct {
	Dir  *os.File
	Path string

	clients      ClientMap
	clientsMutex sync.RWMutex

	cmdgrp *exec.Group

	dbg Debugger

	sessionClients      map[string]*sessionClient
	sessionClientsMutex sync.RWMutex
}

// NewSession creates a new session.
func NewSession(ctx context.Context, dbg Debugger, file string) (*Session, error) {
	s := &Session{
		Path:           file,
		clients:        ClientMap{},
		cmdgrp:         exec.NewGroup(ctx),
		dbg:            dbg,
		sessionClients: map[string]*sessionClient{},
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

// Logs returns a channel that emits lines from the log file of a given client.
// An error will be returned if the client with the provided name does not exist
// or if logtype is not stderr or stdout.
func (s *Session) Logs(clientName string, fd int32) (osc.Message, error) {
	clientPath := filepath.Join(s.Path, clientName)
	var (
		client *sessionClient
		exists bool
	)
	s.clientsMutex.RLock()
	client, exists = s.sessionClients[clientPath]
	s.clientsMutex.RUnlock()
	if !exists {
		return osc.Message{}, errors.New("client does not exist: " + clientName)
	}
	if fd != stdoutArg && fd != stderrArg {
		return osc.Message{}, errors.Errorf("fd must be either %d or %d", stdoutArg, stderrArg)
	}

	var streamPath string
	if fd == stdoutArg {
		s.dbg.Debugf("returning logs from stdout stream")
		streamPath = client.stdoutPath
	} else {
		s.dbg.Debugf("returning logs from stderr stream")
		streamPath = client.stderrPath
	}
	f, err := os.Open(streamPath)
	if err != nil {
		return osc.Message{}, errors.Wrapf(err, "opening %s", streamPath)
	}

	s.dbg.Debugf("getting logs from file %s", streamPath)

	return s.linesToMessage(f, clientName)
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
	var (
		clientPath = filepath.Join(s.Path, cmdname)
		stdoutPath = filepath.Join(clientPath, stdoutFilename)
		stderrPath = filepath.Join(clientPath, stderrFilename)
	)
	s.dbg.Debugf("client stdout path %s", stdoutPath)
	s.dbg.Debugf("client stderr path %s", stderrPath)

	// Create the files where we will store the output.
	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return errors.Wrap(err, "creating "+stdoutPath)
	}
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return errors.Wrap(err, "creating "+stderrPath)
	}

	// Pipe the output to the newly created files.
	stdout, stderr, err := s.cmdgrp.Output(cmdname)
	if err != nil {
		return errors.Wrap(err, "getting output for "+cmdname)
	}
	g.Go(s.pipeSync(stdoutFile, stdout))
	g.Go(s.pipeSync(stderrFile, stderr))

	// Add the pipes to the session clients.
	s.sessionClientsMutex.Lock()
	if _, ok := s.sessionClients[clientPath]; !ok {
		s.sessionClients[clientPath] = &sessionClient{}
	}
	s.sessionClients[clientPath].stderrPath = stderrPath
	s.sessionClients[clientPath].stdoutPath = stdoutPath
	s.sessionClientsMutex.Unlock()

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

	// Get the arguments.
	cmdname, err := msg.Arguments[0].ReadString()
	if err != nil {
		return "", errors.Wrap(err, "could not read cmdname")
	}
	progname, err := msg.Arguments[1].ReadString()
	if err != nil {
		return "", errors.Wrap(err, "could not read progname")
	}

	// Exec the new client.
	var (
		cmd       = exec.Command(progname)
		localAddr = local.LocalAddr().String()
	)
	cmd.Env = append(os.Environ(), "NSM_URL="+localAddr)

	if err := s.cmdgrp.AddCmd(cmdname, cmd); err != nil {
		return "", errors.Wrap(err, "adding command "+progname)
	}

	// Create a new entry in the session clients map.
	s.sessionClientsMutex.Lock()
	s.sessionClients[filepath.Join(s.Path, cmdname)] = &sessionClient{}
	s.sessionClientsMutex.Unlock()

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

// pipeSync pipes an io.ReadCloser to a file and calls
// Sync on the *os.File after every write.
func (s *Session) pipeSync(fd *os.File, r io.ReadCloser) func() error {
	return func() error {
		for {
			buf := make([]byte, 256)
			if _, err := r.Read(buf); err != nil {
				if err == io.EOF {
					break
				}
			}
			s.dbg.Debugf("writing %s to %s", string(buf), fd.Name())
			if _, err := fd.Write(buf); err != nil {
				return errors.Wrap(err, "writing to file")
			}
			if err := fd.Sync(); err != nil {
				return errors.Wrap(err, "syncing file")
			}
		}
		return nil
	}
}

// linesToMessage converts lines from the provided io.Reader to an OSC message.
func (s *Session) linesToMessage(r io.Reader, clientName string) (osc.Message, error) {
	var (
		br    = bufio.NewScanner(r)
		lines = []string{}
		m     = osc.Message{
			Address: nsm.AddressReply,
			Arguments: osc.Arguments{
				osc.String(nsm.AddressClientLogs),
				osc.String(clientName),
			},
		}
	)
	for br.Scan() {
		txt := strings.TrimSpace(strings.Trim(br.Text(), "\x00"))
		s.dbg.Debug("got output " + txt + " for client " + clientName)
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	if err := br.Err(); err != nil {
		return osc.Message{}, errors.Wrap(br.Err(), "scanning lines to channel")
	}
	m.Arguments = append(m.Arguments, osc.Int(len(lines)))

	for _, line := range lines {
		m.Arguments = append(m.Arguments, osc.String(line))
	}
	return m, nil
}
