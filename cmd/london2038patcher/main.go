package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
	"github.com/ricochhet/london2038patcher/pkg/dlutil"
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

	cmd, err := commands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
	}

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

	_ = patcher.downloadCmd()
}

// commands handles the specified command flags.
func commands() (bool, error) {
	if flag.NArg() == 0 {
		return false, nil
	}

	cmd := strings.ToLower(flag.Args()[0])
	args := flag.Args()[1:]

	lr := patchutil.NewDefaultLocaleRegistry()

	lf, err := patchutil.NewLocaleFilter(lr, toSlice(Flag.Locales, ","))
	if err != nil {
		return true, err
	}

	d := patchutil.Options{
		Registry: lr,
		Filter:   lf,
		IdxOptions: &patchutil.IdxOptions{
			CRC32: Flag.CRC32,
		},
		Archs: toSlice(Flag.Archs, ","),
	}

	switch cmd {
	case "decodeidx":
		check(false, 3)
		return true, decodeCmd(args...)
	case "encodeidx":
		check(false, 3)
		return true, encodeCmd(args...)
	case "unpack":
		check(false, 4)
		return true, unpackCmd(d, args...)
	case "pack":
		check(false, 4)
		return true, packCmd(d, args...)
	case "packwithidx":
		check(false, 4)
		return true, packWithIdxCmd(d, args...)
	case "unpackfromfile":
		check(false, 3)
		return true, unpackFromFileCmd(d, args...)
	case "regedit":
		check(true, 3)
		return true, regeditCmd(args...)
	case "help", "h":
		usage()
	}

	if winutil.IsAdmin() {
		cmdutil.Pause()
	}

	return false, nil
}

// check handles checks for commands.
func check(canBeUnsupported bool, v int) {
	if canBeUnsupported {
		maybeUnsupported()
	}

	if flag.NArg() < v {
		usage()
	}
}

// maybeUnsupported exits with code 1 if the current runtime is not Windows.
func maybeUnsupported() {
	if runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "This command is unsupported on non-Windows machines.\n")
		os.Exit(1)
	}
}
