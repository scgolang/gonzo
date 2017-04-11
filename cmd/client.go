package cmd

import (
	"net"

	"github.com/scgolang/nsm"
)

// Client represents an nsm client.
// Note that this is the information that is provided to us
// by clients when they announce.
// There is also other information about clients that we store directly on the session.
type Client struct {
	Addr            net.Addr         `json:"addr"`
	ApplicationName string           `json:"application_name"`
	Capabilities    nsm.Capabilities `json:"capabilities"`
	ExecutableName  string           `json:"executable_name"`
	Major           int32            `json:"major"`
	Minor           int32            `json:"minor"`
}

// Pid is a process ID.
type Pid int32

// ClientMap helps track clients by process ID.
type ClientMap map[Pid]*Client
