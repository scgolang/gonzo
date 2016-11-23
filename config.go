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
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`
}

// NewConfig creates a new config from command line flags.
func NewConfig() (Config, error) {
	c := Config{}
	flag.StringVar(&c.Host, "h", "127.0.0.1", "host")
	flag.IntVar(&c.Port, "p", DefaultPort, "port")
	flag.BoolVar(&c.Debug, "debug", false, "Print debugging output")
	flag.Parse()
	return c, nil
}
