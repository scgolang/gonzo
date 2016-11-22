package main

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// App contains all the state for the application.
type App struct {
	Config
	errgroup.Group
	osc.Conn
}

// NewApp creates a new application.
func NewApp(config Config) (*App, error) {
	app := &App{Config: config}
	if err := app.initialize(); err != nil {
		return nil, errors.Wrap(err, "could not initialize application")
	}
	return app, nil
}

// initialize initializes the application.
func (app *App) initialize() error {
	// Initialize the osc listener.
	listenAddr := net.JoinHostPort(app.Host, strconv.Itoa(app.Port))
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return errors.Wrap(err, "could not resolve udp address")
	}
	conn, err := osc.ListenUDP("udp", addr)
	if err != nil {
		return errors.Wrap(err, "could not listen on udp")
	}
	app.Conn = conn
	app.Go(app.ServeOSC)
	return nil
}

// ServeOSC serves osc requests.
func (app *App) ServeOSC() error {
	return app.Serve(app.dispatcher())
}

// dispatcher returns the osc Dispatcher for the application.
func (app *App) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		nsm.AddressServerAdd:      app.Add,
		nsm.AddressServerAnnounce: app.Announce,
		nsm.AddressServerList:     app.ListProjects,
		nsm.AddressReply:          app.Reply,
	}
}

// Announce announces new clients.
func (app *App) Announce(msg *osc.Message) error {
	return nil
}

// Reply handles replies from clients.
func (app *App) Reply(msg *osc.Message) error {
	return nil
}
