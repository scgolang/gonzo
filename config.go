package main

import (
	"flag"
)

// Config provides configuration for the application.
type Config struct {
	Host string
	Port int
}

// NewConfig creates a new config from command line flags.
func NewConfig() (Config, error) {
	c := Config{}
	flag.StringVar(&c.Host, "h", "0.0.0.0", "host")
	flag.IntVar(&c.Port, "p", 0, "port")
	flag.Parse()
	return c, nil
}
