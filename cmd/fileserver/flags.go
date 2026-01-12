package main

import (
	"flag"
)

type Flags struct {
	Version    bool
	ConfigFile string

	CertFile string
	KeyFile  string
}

var Flag = NewFlags()

// NewFlags creates an empty Flags.
func NewFlags() *Flags {
	return &Flags{}
}

//nolint:gochecknoinits // wontfix
func init() {
	registerFlags(flag.CommandLine, Flag)
	flag.Parse()
}

// registerFlags registers all flags to the flagset.
func registerFlags(fs *flag.FlagSet, f *Flags) {
	fs.StringVar(&f.ConfigFile, "c", "fileserver.json", "Path to file server configuration")
	fs.StringVar(&f.CertFile, "cert", "", "TLS cert")
	fs.StringVar(&f.KeyFile, "key", "", "TLS key")
	fs.BoolVar(&f.Version, "version", false, "Show version information")
}
