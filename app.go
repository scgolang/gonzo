package main

import (
	"context"
	"log"
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
	osc.Conn

	Announcements chan osc.Message
	Errors        chan osc.Message
	Replies       chan osc.Message

	Capabilities nsm.Capabilities

	ctx      context.Context
	errgrp   *errgroup.Group
	sessions *Sessions
}

// NewApp creates a new application.
func NewApp(ctx context.Context, config Config) (*App, error) {
	g, gctx := errgroup.WithContext(ctx)

	sessions, err := NewSessions(config.Home, gctx)
	if err != nil {
		return nil, errors.Wrap(err, "opening sessions")
	}

	app := &App{
		Config: config,

		Announcements: make(chan osc.Message),
		Errors:        make(chan osc.Message),
		Replies:       make(chan osc.Message),

		Capabilities: nsm.Capabilities{nsm.CapServerControl},

		ctx:      gctx,
		errgrp:   g,
		sessions: sessions,
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

// OscMethod returns an osc.Method which is based on an NsmMethod.
func (app *App) OscMethod(method NsmMethod, addr string) osc.Method {
	return func(msg osc.Message) error {
		var reply osc.Message

		message, err := method(msg)
		if err != nil {
			reply = ReplyError(addr, err.Code(), err.Error())
		} else {
			reply = ReplySuccess(msg.Sender, addr, message)
		}
		return errors.Wrap(app.SendTo(msg.Sender, reply), "sending reply")
	}
}

// Ping handles /ping messages
func (app *App) Ping(msg osc.Message) error {
	return errors.Wrap(app.SendTo(msg.Sender, osc.Message{Address: "/pong"}), "sending pong")
}

// Reply handles replies from clients.
func (app *App) Reply(msg osc.Message) error {
	return nil
}

// ServeOSC serves osc requests.
func (app *App) ServeOSC() error {
	return app.Serve(app.dispatcher())
}

// Wait waits for all the goroutines to return nil, or for one of them to return a non-nil value, whichever happens first.
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
		nsm.AddressServerAnnounce: app.OscMethod(app.Announce, nsm.AddressServerAnnounce),
		nsm.AddressServerClients:  app.ListClients,
		nsm.AddressServerSessions: app.ListSessions,
		nsm.AddressServerNew:      app.OscMethod(app.NewSession, nsm.AddressServerNew),
		"/ping":                   app.Ping,
		nsm.AddressServerRemove:   app.OscMethod(app.RemoveSession, nsm.AddressServerRemove),
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

// NsmMethod is a utility type that is used by osc methods that should
// always generate an nsm-style reply to the client.
type NsmMethod func(msg osc.Message) (string, nsm.Error)

// ReplyError returns the message used to signal an error to a client.
func ReplyError(address string, code nsm.Code, message string) osc.Message {
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
func ReplySuccess(remote net.Addr, address, message string) osc.Message {
	return osc.Message{
		Address: nsm.AddressReply,
		Arguments: osc.Arguments{
			osc.String(address),
			osc.String(message),
		},
	}
}
