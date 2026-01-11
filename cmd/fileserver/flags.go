package main

import (
	"flag"
)

type Flags struct {
	Version bool
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
	fs.BoolVar(&f.Version, "version", false, "Show version information")
}
