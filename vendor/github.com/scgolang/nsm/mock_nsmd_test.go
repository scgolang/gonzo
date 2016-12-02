package nsm

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

type mockNsmdConfig struct {
	listenAddr string

	announcePause time.Duration
	announceReply osc.Message
}

// mockNsmd mocks an nsmd server.
type mockNsmd struct {
	mockNsmdConfig
	osc.Conn
	errgroup.Group

	t             *testing.T
	timeout       time.Duration
	clientAddr    net.Addr
	openChan      chan osc.Message
	saveChan      chan osc.Message
	announceAcked chan struct{}
}

// newMockNsmd creates a new mock nsmd server.
// This has the side effect of setting the NSM_URL to the listening
// address of the mock server.
func newMockNsmd(t *testing.T, config mockNsmdConfig) *mockNsmd {
	nsmd := &mockNsmd{
		mockNsmdConfig: config,
		t:              t,
		announceAcked:  make(chan struct{}),
		openChan:       make(chan osc.Message),
		saveChan:       make(chan osc.Message),
	}
	nsmd.defaults()
	nsmd.initialize()
	return nsmd
}

func (m *mockNsmd) defaults() {
	if m.announceReply.Address == "" { // There is no announce reply.
		m.announceReply = osc.Message{
			Address: AddressReply,
			Arguments: osc.Arguments{
				osc.String(AddressServerAnnounce),
				osc.String("session started"),
				osc.String("mock_nsmd"),
				osc.String(Capabilities{CapServerBroadcast, CapServerControl}.String()),
			},
		}
	}
	if m.announcePause == time.Duration(0) {
		m.announcePause = 10 * time.Millisecond
	}
	if m.timeout == time.Duration(0) {
		m.timeout = 4 * time.Second
	}
}

func (m *mockNsmd) initialize() {
	laddr, err := net.ResolveUDPAddr("udp", m.listenAddr)
	if err != nil {
		m.t.Fatalf("could not resolve udp address: %s", err)
	}
	conn, err := osc.ListenUDP("udp", laddr)
	if err != nil {
		m.t.Fatalf("could not create udp connection: %s", err)
	}
	if err := os.Setenv(NsmURL, conn.LocalAddr().String()); err != nil {
		m.t.Fatalf("could not set %s environment variable: %s", NsmURL, err)
	}
	m.Conn = conn

	m.Go(m.startOSC)
}

// OpenSession sends the provided message (which may or may not be an open message),
// and waits for a reply with a configurable timeout.
func (m *mockNsmd) OpenSession(msg osc.Message) osc.Message {
	return m.serverToClient("open", m.openChan, msg)
}

// OpenSession sends the provided message (which may or may not be an open message),
// and waits for a reply with a configurable timeout.
func (m *mockNsmd) SaveSession(msg osc.Message) osc.Message {
	return m.serverToClient("save", m.saveChan, msg)
}

// SessionLoaded triggers a session_is_loaded message.
func (m *mockNsmd) SessionLoaded() {
	<-m.announceAcked

	if err := m.SendTo(m.clientAddr, osc.Message{Address: AddressClientSessionIsLoaded}); err != nil {
		m.t.Fatalf("SessionLoaded: %s", err)
	}
}

// Close closes the mock server.
func (m *mockNsmd) Close() error {
	close(m.openChan)
	close(m.saveChan)

	errs := []string{}
	if err := os.Unsetenv(NsmURL); err != nil {
		errs = append(errs, err.Error())
	}
	if err := m.Conn.Close(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, " and "))
}

func (m *mockNsmd) serverToClient(cmdName string, replyChan chan osc.Message, msg osc.Message) osc.Message {
	// Wait for announcement from the client.
	<-m.announceAcked

	mc := make(chan osc.Message)
	go func(mc chan osc.Message) {
		var (
			timeout = time.After(m.timeout)
			reply   osc.Message
		)
		select {
		case reply = <-replyChan:
			mc <- reply
		case <-timeout:
			m.t.Fatalf("timeout after %s waiting for %s reply", m.timeout.String(), cmdName)
		}
		mc <- osc.Message{}
	}(mc)

	if err := m.SendTo(m.clientAddr, msg); err != nil {
		m.t.Fatalf("sending %s message: %s", cmdName, err)
	}
	return <-mc
}

func (m *mockNsmd) startOSC() error {
	return m.Serve(m.dispatcher())
}

func (m *mockNsmd) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		AddressServerAnnounce: func(msg osc.Message) error {
			m.clientAddr = msg.Sender

			// Sleep then reply.
			time.Sleep(m.announcePause)
			if err := m.SendTo(msg.Sender, m.announceReply); err != nil {
				return err
			}
			close(m.announceAcked)
			return nil
		},
		AddressReply: func(msg osc.Message) error {
			if len(msg.Arguments) < 1 {
				return errors.New("/reply must provide the address being replied to")
			}
			addr, err := msg.Arguments[0].ReadString()
			if err != nil {
				return errors.Wrap(err, "expected first argument to be a string")
			}
			switch addr {
			case AddressClientOpen:
				m.openChan <- msg
			case AddressClientSave:
				m.saveChan <- msg
			}
			return nil
		},
	}
}

type mockReply struct {
	Message string
	Err     Error
}

type mockSession struct {
	SessionInfo

	open mockReply
	save mockReply

	// Note that client code needs to initialize these.
	loadedChan  chan struct{}
	showGuiChan chan bool
}

func (m *mockSession) Open(info SessionInfo) (string, Error) {
	m.SessionInfo = info
	return m.open.Message, m.open.Err
}

func (m *mockSession) Save() (string, Error) {
	return m.save.Message, m.save.Err
}

func (m *mockSession) IsLoaded() error {
	close(m.loadedChan)
	return nil
}

func (m *mockSession) ShowGUI(show bool) error {
	m.showGuiChan <- show
	return nil
}
