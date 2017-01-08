package main

import (
	"flag"
	"os"
	"path/filepath"
)

const (
	// DefaultPort is the default listening port.
	DefaultPort = 56070
)

// Config provides configuration for the application.
type Config struct {
	Home      string `json:"home"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	DebugFlag bool   `json:"debug"`
}

// NewConfig creates a new config from command line flags.
func NewConfig() (Config, error) {
	var (
		c           = Config{}
		defaultHome = filepath.Join(os.Getenv("HOME"), "gonzo-sessions")
	)
	flag.StringVar(&c.Home, "home", defaultHome, "Session manager's home directory")
	flag.StringVar(&c.Host, "h", "127.0.0.1", "host")
	flag.IntVar(&c.Port, "p", DefaultPort, "port")
	flag.BoolVar(&c.DebugFlag, "debug", false, "Print debugging output")
	flag.Parse()
	return c, nil
}
