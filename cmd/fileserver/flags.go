package main

import (
	"flag"

	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
)

type Flags struct {
	ConfigFile string

	CertFile string
	KeyFile  string
}

var (
	flags = NewFlags()
	cmds  = cmdutil.Commands{
		{Usage: "fileserver help", Desc: "Show this help"},
		{Usage: "fileserver list [PATH]", Desc: "List embedded files"},
		{Usage: "fileserver dump [PATH]", Desc: "Dump embedded files to disk"},
		{Usage: "fileserver version", Desc: "Display fileserver version"},
	}
)

// NewFlags creates an empty Flags.
func NewFlags() *Flags {
	return &Flags{}
}

//nolint:gochecknoinits // wontfix
func init() {
	registerFlags(flag.CommandLine, flags)
	flag.Parse()
}

// registerFlags registers all flags to the flagset.
func registerFlags(fs *flag.FlagSet, f *Flags) {
	fs.StringVar(&f.ConfigFile, "c", "fileserver.json", "Path to file server configuration")
	fs.StringVar(&f.CertFile, "cert", "", "TLS cert")
	fs.StringVar(&f.KeyFile, "key", "", "TLS key")
}
