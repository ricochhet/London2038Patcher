package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/regutil"
	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
	"github.com/ricochhet/london2038patcher/pkg/winutil"
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
		HTTPClient:    *dlutil.NewHTTPClient(time.Duration(Flag.Timeout)),
		ChecksumURL:   Flag.ChecksumURL,
		PatchURL:      Flag.PatchURL,
		ChecksumFile:  Flag.ChecksumFile,
		HellgateCUKey: "",
		HellgateKey:   "",
		UsePatchDir:   Flag.PatchDir,
		patchDir:      "",
	})

	_ = patcher.download()
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
	case "regedit":
		maybeUnsupported()

		if flag.NArg() < 3 {
			usage()
		}

		_ = regedit(args...)

		return true
	case "help", "h":
		usage()
	}

	if winutil.IsAdmin() {
		cmdutil.Pause()
	}

	return false
}

// download command.
func (p *PatcherCtx) download() error {
	return timeutil.Timer(func() error {
		pErr := p.Get().Download()
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

// regedit command.
func regedit(a ...string) error {
	return timeutil.Timer(func() error {
		rErr := regutil.Regedit(a[0], a[1])
		if rErr != nil {
			fmt.Fprintf(os.Stderr, "Error editing registry: %v\n", rErr)
		}

		return rErr
	}, "Regedit", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// maybeUnsupported exits with code 1 if the current runtime is not Windows.
func maybeUnsupported() {
	if runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "This command is unsupported on non-Windows machines.\n")
		os.Exit(1)
	}
}
