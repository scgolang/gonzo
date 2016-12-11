package main

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/scgolang/exec"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// App contains all the state for the application.
type App struct {
	Config
	osc.Conn

	Announcements chan osc.Message
	Errors        chan osc.Message
	Replies       chan osc.Message

	clients ClientMap
	cmdgrp  *exec.CmdGroup
	ctx     context.Context
	errgrp  *errgroup.Group
}

// NewApp creates a new application.
func NewApp(ctx context.Context, config Config) (*App, error) {
	g, gctx := errgroup.WithContext(ctx)

	app := &App{
		Config: config,

		Announcements: make(chan osc.Message),
		Errors:        make(chan osc.Message),
		Replies:       make(chan osc.Message),

		clients: ClientMap{},
		cmdgrp:  exec.NewCmdGroup(gctx),
		ctx:     gctx,
		errgrp:  g,
	}
	if err := app.initialize(); err != nil {
		return nil, errors.Wrap(err, "could not initialize application")
	}
	return app, nil
}

// Go runs a new goroutine as part of an errgroup.Group
func (app *App) Go(f func() error) {
	app.errgrp.Go(f)
}

// Ping handles /ping messages
func (app *App) Ping(msg osc.Message) error {
	return errors.Wrap(app.SendTo(msg.Sender, osc.Message{Address: "/pong"}), "sending pong")
}

// Reply handles replies from clients.
func (app *App) Reply(msg osc.Message) error {
	return nil
}

// ReplyError returns the message used to signal an error to a client.
func (app *App) ReplyError(address string, code nsm.Code, message string) osc.Message {
	return osc.Message{
		Address: nsm.AddressError,
		Arguments: osc.Arguments{
			osc.String(address),
			osc.Int(code),
			osc.String(message),
		},
	}
}

// ReplySuccess returns the message used to signal a successful operation.
func (app *App) ReplySuccess(remote net.Addr, address string) osc.Message {
	return osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(address),
		},
	}
}

// ServeOSC serves osc requests.
func (app *App) ServeOSC() error {
	return app.Serve(app.dispatcher())
}

// Wait waits for all the goroutines to return nil, or for one of them
// to return a non-nil value, whichever happens first.
func (app *App) Wait() error {
	return app.errgrp.Wait()
}

// debug prints a debug message.
func (app *App) debug(msg string) {
	if app.Debug {
		log.Println(msg)
	}
}

// debugf prints a debug message with printf semantics.
func (app *App) debugf(format string, args ...interface{}) {
	if app.Debug {
		log.Printf(format, args...)
	}
}

// dispatcher returns the osc Dispatcher for the application.
func (app *App) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		nsm.AddressServerAdd:      app.Add,
		nsm.AddressServerClients:  app.ListClients,
		nsm.AddressServerAnnounce: app.Announce,
		nsm.AddressServerProjects: app.ListProjects,
		"/ping":                   app.Ping,
		nsm.AddressReply:          app.Reply,
	}
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
