package main

import (
	"flag"
	"strings"
)

type Flags struct {
	ChecksumURL  string
	PatchURL     string
	ChecksumFile string
	PatchDir     bool
	Version      bool
	Timeout      int
	Locales      string
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
	fs.StringVar(
		&f.ChecksumURL,
		"checksum-url",
		"https://auth.london2038.com/patcher/checksums.xml",
		"URL for checksum file",
	)
	fs.StringVar(
		&f.PatchURL,
		"patch-url",
		"https://auth.london2038.com/patcher/",
		"URL for patch files",
	)
	fs.StringVar(
		&f.ChecksumFile,
		"checksum-file",
		"checksums.xml",
		"Path to save checksum file to",
	)
	fs.BoolVar(&f.PatchDir, "patch-dir", false, "Use patch directory for files")
	fs.BoolVar(&f.Version, "version", false, "Show version information")
	fs.IntVar(&f.Timeout, "timeout", 0, "Set download timeout")
	fs.StringVar(&f.Locales, "locales", "en", "Set locale code for un/packing")
}

// toSlice converts a string to a slice.
func toSlice(s, sep string) []string {
	return strings.Split(s, sep)
}
