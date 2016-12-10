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

	t          *testing.T
	timeout    time.Duration
	clientAddr net.Addr

	announceAcked chan struct{}

	openChan chan osc.Message
	saveChan chan osc.Message
	openErr  chan Error
	saveErr  chan Error

	dirtyChan      chan bool
	guiShowingChan chan bool
	progressChan   chan float32
	statusChan     chan ClientStatus
}

// newMockNsmd creates a new mock nsmd server.
// This has the side effect of setting the NSM_URL to the listening
// address of the mock server.
func newMockNsmd(t *testing.T, config mockNsmdConfig) *mockNsmd {
	nsmd := &mockNsmd{
		mockNsmdConfig: config,

		t: t,

		announceAcked: make(chan struct{}),

		openChan: make(chan osc.Message),
		saveChan: make(chan osc.Message),
		openErr:  make(chan Error),
		saveErr:  make(chan Error),

		dirtyChan:      make(chan bool),
		guiShowingChan: make(chan bool),
		progressChan:   make(chan float32),
		statusChan:     make(chan ClientStatus),
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

// OpenSessionError sends the provided message (which may or may not be an open message),
// and waits for an error with a configurable timeout.
func (m *mockNsmd) OpenSessionError(msg osc.Message) Error {
	return m.serverToClientError("open", m.openErr, msg)
}

// OpenSession sends the provided message (which may or may not be an open message),
// and waits for a reply with a configurable timeout.
func (m *mockNsmd) SaveSession(msg osc.Message) osc.Message {
	return m.serverToClient("save", m.saveChan, msg)
}

// SessionLoaded triggers a session_is_loaded server to client message.
func (m *mockNsmd) SessionLoaded() {
	<-m.announceAcked

	if err := m.SendTo(m.clientAddr, osc.Message{Address: AddressClientSessionIsLoaded}); err != nil {
		m.t.Fatalf("SessionLoaded: %s", err)
	}
}

// HideOptionalGUI triggers a hide_optional_gui server to client message.
func (m *mockNsmd) HideOptionalGUI() {
	<-m.announceAcked

	if err := m.SendTo(m.clientAddr, osc.Message{Address: AddressClientHideOptionalGUI}); err != nil {
		m.t.Fatalf("SessionLoaded: %s", err)
	}
}

// ShowOptionalGUI triggers a show_optional_gui server to client message.
func (m *mockNsmd) ShowOptionalGUI() {
	<-m.announceAcked

	if err := m.SendTo(m.clientAddr, osc.Message{Address: AddressClientShowOptionalGUI}); err != nil {
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

func (m *mockNsmd) serverToClientError(cmdName string, errChan chan Error, msg osc.Message) Error {
	// Wait for announcement from the client.
	<-m.announceAcked

	ec := make(chan Error)
	go func(ec chan Error) {
		var (
			timeout = time.After(m.timeout)
			nsmErr  Error
		)
		select {
		case nsmErr = <-errChan:
			ec <- nsmErr
		case <-timeout:
			m.t.Fatalf("timeout after %s waiting for %s reply", m.timeout.String(), cmdName)
		}
		ec <- nil
	}(ec)

	if err := m.SendTo(m.clientAddr, msg); err != nil {
		m.t.Fatalf("sending %s message: %s", cmdName, err)
	}
	return <-ec
}

func (m *mockNsmd) startOSC() error {
	return m.Serve(m.dispatcher())
}

func (m *mockNsmd) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		AddressClientProgress: m.ProgressHandler,
		AddressClientStatus:   m.ClientStatusHandler,
		AddressError:          m.ErrorHandler,
		AddressReply:          m.ReplyHandler,
		AddressServerAnnounce: m.AnnounceHandler,

		AddressClientGUIHidden: func(msg osc.Message) error {
			m.guiShowingChan <- false
			return nil
		},
		AddressClientGUIShowing: func(msg osc.Message) error {
			m.guiShowingChan <- true
			return nil
		},
		AddressClientIsDirty: func(msg osc.Message) error {
			m.dirtyChan <- true
			return nil
		},
		AddressClientIsClean: func(msg osc.Message) error {
			m.dirtyChan <- false
			return nil
		},
	}
}

func (m *mockNsmd) AnnounceHandler(msg osc.Message) error {
	m.clientAddr = msg.Sender

	// Sleep then reply.
	time.Sleep(m.announcePause)
	if err := m.SendTo(msg.Sender, m.announceReply); err != nil {
		return err
	}
	close(m.announceAcked)
	return nil
}

func (m *mockNsmd) ClientStatusHandler(msg osc.Message) error {
	if len(msg.Arguments) != 2 {
		return errors.New(AddressClientStatus + " should have exactly two arguments")
	}
	priority, err := msg.Arguments[0].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "could not read priority from "+AddressClientStatus)
	}
	statusMsg, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "could not read status message from "+AddressClientStatus)
	}
	m.statusChan <- ClientStatus{Priority: int(priority), Message: statusMsg}
	return nil
}

func (m *mockNsmd) ErrorHandler(msg osc.Message) error {
	if len(msg.Arguments) < 3 {
		return errors.New("/error must have at least 3 arguments")
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading first argument")
	}
	code, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading second argument")
	}
	errmsg, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading third argument")
	}
	nsmErr := NewError(Code(code), errmsg)
	switch addr {
	case AddressClientOpen:
		m.openErr <- nsmErr
	case AddressClientSave:
		m.saveErr <- nsmErr
	}
	return nil
}

func (m *mockNsmd) ProgressHandler(msg osc.Message) error {
	if len(msg.Arguments) != 1 {
		return errors.New(AddressClientProgress + " should have exactly one argument")
	}
	x, err := msg.Arguments[0].ReadFloat32()
	if err != nil {
		return errors.Wrap(err, "read first argument in "+AddressClientProgress+" message")
	}
	m.progressChan <- x
	return nil
}

func (m *mockNsmd) ReplyHandler(msg osc.Message) error {
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
}

type mockReply struct {
	Message string
	Err     Error
}

type mockSession struct {
	SessionInfo

	open mockReply
	save mockReply

	// Note that client code needs to initialize the channels below.

	// Server to client requests.
	loadedChan  chan struct{}
	showGuiChan chan bool

	// Client to server informational messages.
	dirtyChan      chan bool
	progressChan   chan float32
	showingGuiChan chan bool
	statusChan     chan ClientStatus

	// Extra osc methods to add to the client.
	methods osc.Dispatcher
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

func (m *mockSession) Dirty() chan bool {
	return m.dirtyChan
}

func (m *mockSession) GUIShowing() chan bool {
	return m.showingGuiChan
}

func (m *mockSession) Progress() chan float32 {
	return m.progressChan
}

func (m *mockSession) ClientStatus() chan ClientStatus {
	return m.statusChan
}

func (m *mockSession) Methods() osc.Dispatcher {
	return m.methods
}
