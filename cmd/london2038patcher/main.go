package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/patchutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() string {
	return fmt.Sprintf(
		"London2038Patcher\n\tBuild Date: %s\n\tGit Hash: %s\n\tBuilt On: %s\n",
		buildDate, gitHash, buildOn,
	)
}

func usage() {
	flag.Usage()
	os.Exit(0)
}

func main() {
	if Flag.Version {
		fmt.Fprint(os.Stdout, version())
		return
	}

	cmd := commands()
	if cmd {
		return
	}

	patcher := NewPatcherCtx()
	patcher.Set(&Patcher{
		HTTPClient:   *dlutil.NewHTTPClient(time.Duration(Flag.Timeout)),
		ChecksumURL:  Flag.ChecksumURL,
		PatchURL:     Flag.PatchURL,
		ChecksumFile: Flag.ChecksumFile,
		// PatchDir:     "",
	})

	_ = download(patcher)
}

// commands handles the specified command flags.
func commands() bool {
	if flag.NArg() == 0 {
		return false
	}

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "unpack":
		if flag.NArg() < 4 {
			usage()
		}

		_ = unpack(args...)

		return true
	case "help", "h":
		usage()
	}

	return false
}

// download command.
func download(patcher *PatcherCtx) error {
	return timeutil.Timer(func() error {
		pErr := patcher.Get().Download()
		if pErr != nil {
			fmt.Fprintf(os.Stderr, "Error downloading files: %v\n", pErr)
		}

		return pErr
	}, "Download", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// unpack command.
func unpack(a ...string) error {
	return timeutil.Timer(func() error {
		uErr := patchutil.Unpacker(a[0], a[1], a[2])
		if uErr != nil {
			fmt.Fprintf(os.Stderr, "Error unpacking patch: %v\n", uErr)
		}

		return uErr
	}, "Unpack", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}
