package main

import (
	"flag"
)

const (
	// DefaultPort is the default listening port.
	DefaultPort = 56070
)

// Config provides configuration for the application.
type Config struct {
	Host string
	Port int
}

// NewConfig creates a new config from command line flags.
func NewConfig() (Config, error) {
	c := Config{}
	flag.StringVar(&c.Host, "h", "127.0.0.1", "host")
	flag.IntVar(&c.Port, "p", DefaultPort, "port")
	flag.Parse()
	return c, nil
}
