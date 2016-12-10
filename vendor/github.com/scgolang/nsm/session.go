package nsm

import (
	"net"

	"github.com/scgolang/osc"
)

// Session represents the behavior of a client
// with respect to the control messages that are
// sent by Non Session Manager.
// The only methods in this interface that are actually
// REQUIRED by all implementations are Open and Save.
// The other methods are all optional and mostly depend on
// the capabilities that a given client wishes to implement
// (e.g. progress reports for long-running open and save operations).
type Session interface {
	// Announce will be called when the server has replied
	// to the client's announce message.
	// The server sends back information about itself in the reply.
	// This method is optional if your Session implementation embeds SessionInfo.
	// The SessionInfo implementation automatically assigns the ServerInfo
	// passed to Announce as the Server field on the your struct.
	Announce(ServerInfo) error

	// Open tells the client to open a session.
	// If a client has not specified CapSwitch in their
	// capabilities then this method will only be called once.
	Open(SessionInfo) (string, Error)

	// Save will only be called after Open has been
	// called, and may be called any number of times during
	// a session. If the user aborts the session this method
	// will not be called.
	Save() (string, Error)

	// IsLoaded will be invoked when Non Session Manager
	// has started all the clients for a given session.
	IsLoaded() error

	// ShowGUI is an optional method that may be invoked in response
	// to server commands. If your nsm client has the CapGUI capability
	// then it should show/hide the GUI based on the bool parameter.
	ShowGUI(bool) error

	// Dirty is an optional client to server informational message
	// that tells Non Session Manager that a client has unsaved changes.
	// If your client does not support CapDirty then return nil from
	// this method.
	Dirty() chan bool

	// GUIShowing should return a channel that is used to notify
	// Non Session Manager that the client's GUI is hidden (false)
	// or showing (true). If your client has the CapGUI capability
	// then it must implement this method.
	GUIShowing() chan bool

	// Progress can be used by clients with the CapProgress capability
	// to indicate an ongoing save or open operation.
	// Note that clients with CapProgress are still required to call
	// Open or Save when the operation has completed.
	// The float32 that is sent on this channel must be between 0 and 1,
	// with 1 indicating completion.
	Progress() chan float32

	// ClientStatus can be used by clients with the CapMessage capability
	// to send status updates to Non Session Manager.
	ClientStatus() chan ClientStatus

	// Methods is an optional method that should be used by clients
	// who wish to add their own methods to the OSC server that
	// is used to listen for messages from Non Session Manager.
	// Clients who do not need to listen for OSC messages should return nil.
	// Note that client code should NEVER add a method whose address begins
	// with /nsm since this could damage the communication between
	// the Non Session Manager and the client's application.
	Methods() osc.Dispatcher
}

// SessionInfo contains the data a client receives
// in an Open client control message.
// Note that the optional methods from the Session interface
// are implemented on SessionInfo.
// This means that if your Session implementation embeds
// SessionInfo (which might a good idea anyways because
// your app should probably cache info about the current session)
// you can easily avoid writing boilerplate no-op methods
// for capabilities you do not wish to implement.
type SessionInfo struct {
	// ProjectPath is the path where a client can store
	// their project-specific data.
	// Path can be a directory tree or a file.
	// This is up to the client.
	// If a project exists at the provided path
	// the project must be opened immediately.
	// Otherwise a new project must immediately
	// be created at Path.
	ProjectPath string

	// DisplayName is the name of the client as
	// displayed in Non Session Manager.
	DisplayName string

	// ClientID should be used by clients that expect
	// more than one instance to be part of a single session.
	// ClientID should be prepended to any names it registers
	// with subsystems that could be used by other instances.
	// For example, clients that create JACK connections
	// should prepend ClientID to the JACK client name.
	ClientID string

	// LocalAddr provides the local address of the
	// session's OSC connection.
	LocalAddr net.Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr net.Addr
}

// Announce handles the server's reply to the client's announcement.
func (s SessionInfo) Announce(info ServerInfo) error {
	return nil
}

// IsLoaded is a no-op.
func (s SessionInfo) IsLoaded() error {
	return nil
}

// ShowGUI is a no-op.
func (s SessionInfo) ShowGUI(show bool) error {
	return nil
}

// Dirty returns nil.
func (s SessionInfo) Dirty() chan bool {
	return nil
}

// GUIShowing returns nil.
func (s SessionInfo) GUIShowing() chan bool {
	return nil
}

// Progress returns nil.
func (s SessionInfo) Progress() chan float32 {
	return nil
}

// ClientStatus returns nil.
func (s SessionInfo) ClientStatus() chan ClientStatus {
	return nil
}

// Methods returns an optional osc dispatcher for adding
// endpoints to the OSC server that is used to communicate
// with Non Session Manager.
// This implementation just returns nil.
func (s SessionInfo) Methods() osc.Dispatcher {
	return osc.Dispatcher{}
}

// Message priority levels.
const (
	PriorityNever = iota
	PriorityLow
	PriorityMed
	PriorityHigh
)

// ClientStatus represents a status sent to Non Session Manager.
// Priority should be used to indicate how important it is for
// the user to see the status message, with 0 being lowest priority
// and 3 being highest.
type ClientStatus struct {
	Priority int
	Message  string
}

// Equal returns true if one ClientStatus equals another and false otherwise.
func (status ClientStatus) Equal(other ClientStatus) bool {
	return status.Priority == other.Priority && status.Message == other.Message
}

// ServerInfo contains info about Non Session Manager itself.
// Message is a message that was received in reply to the client's
// announce message. ServerName is the name of the session manager.
// Capabilities describes the capabilities of the session manager.
type ServerInfo struct {
	Message      string
	ServerName   string
	Capabilities Capabilities
}
