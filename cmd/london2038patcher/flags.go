package main

import (
	"flag"

	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
)

type Flags struct {
	QuickEdit bool

	ChecksumURL  string
	PatchURL     string
	ChecksumFile string
	PatchDir     bool
	Timeout      int
	Locales      string
	Archs        string
	CRC32        bool
	Debug        bool
}

var (
	flags = NewFlags()
	cmds  = cmdutil.Commands{
		{Usage: "patcher help", Desc: "Show this help"},
		{Usage: "patcher decodeidx [INDEX] [JSON]", Desc: "Decode an index file into a JSON file"},
		{Usage: "patcher encodeidx [JSON] [INDEX]", Desc: "Decode a JSON file into an index"},
		{
			Usage: "patcher unpack [INDEX] [PATCH] [OUTPUT]",
			Desc:  "Unpack patch into the specified output",
		},
		{
			Usage: "patcher pack [INDEX] [INPUT] [PATCH]",
			Desc:  "Pack input into the specified patch",
		},
		{
			Usage: "patcher packwithidx [INPUT] [INDEX] [PATCH]",
			Desc:  "Pack input into the patch and create an index",
		},
		{
			Usage: "patcher unpackfromfile [JSON] [OUTPUT]",
			Desc:  "Unpack patches specified in JSON file",
		},
		{Usage: "patcher regedit [CU_KEY] [KEY]", Desc: "Add HGL CuKey and Key values to registry"},
		{Usage: "patcher version", Desc: "Display patcher version"},
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
	fs.BoolVar(&f.Debug, "debug", false, "Enable debug mode")
	fs.BoolVar(&f.QuickEdit, "quick-edit", false, "Enable quick edit mode (Windows)")
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
	fs.IntVar(&f.Timeout, "timeout", 0, "Set download timeout")
	fs.StringVar(&f.Locales, "locales", "en", "Set locale code for un/packing")
	fs.StringVar(&f.Archs, "archs", "x64,x86", "Set architectures for un/packing")
	fs.BoolVar(&f.CRC32, "crc32", false, "Hash files with CRC32 when packing files with index")
}
