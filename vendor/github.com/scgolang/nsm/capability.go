package nsm

import (
	"strings"
)

// Capability is a capability of an nsm client.
type Capability string

// CapSep is the separator in the capabilities string
const CapSep = ":"

// Client capabilities.
const (
	CapClientSwitch   Capability = "switch"
	CapClientDirty    Capability = "dirty"
	CapClientProgress Capability = "progress"
	CapClientMessage  Capability = "message"
)

// Server capabilities.
const (
	CapServerControl   Capability = "server_control"
	CapServerBroadcast Capability = "broadcast"
)

// Capabilities shared by the client and the server.
const (
	CapGUI Capability = "optional-gui"
)

// Capabilities represents a list of capabilities
type Capabilities []Capability

// ParseCapabilities parses capabilities from a string.
// Note that this func does not check the capabilities for validity.
func ParseCapabilities(s string) Capabilities {
	caps := Capabilities{}
	s = strings.TrimSuffix(strings.TrimPrefix(s, CapSep), CapSep)
	for _, p := range strings.Split(s, CapSep) {
		caps = append(caps, Capability(p))
	}
	return caps
}

// Equal determines if one set of capabilities matches another.
func (caps Capabilities) Equal(other Capabilities) bool {
	if len(caps) != len(other) {
		return false
	}
	for i, c := range caps {
		if c != other[i] {
			return false
		}
	}
	return true
}

// String converts capabilities to a string.
func (caps Capabilities) String() string {
	if len(caps) == 0 {
		return ""
	}
	ss := make([]string, len(caps))
	for i, cap := range caps {
		ss[i] = string(cap)
	}
	return CapSep + strings.Join(ss, CapSep) + CapSep
}
