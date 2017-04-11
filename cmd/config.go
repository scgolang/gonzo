package cmd

import (
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
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
func NewConfig(args []string, flags *flag.FlagSet) (Config, error) {
	var (
		c           = Config{}
		defaultHome = filepath.Join(os.Getenv("HOME"), "gonzo-sessions")
	)
	flags.StringVar(&c.Home, "home", defaultHome, "Session manager's home directory")
	flags.StringVar(&c.Host, "h", "127.0.0.1", "host")
	flags.IntVar(&c.Port, "p", DefaultPort, "port")
	flags.BoolVar(&c.DebugFlag, "debug", false, "Print debugging output")
	err := flags.Parse(args)
	return c, err
}
