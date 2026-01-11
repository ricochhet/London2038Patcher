package main

import (
	"flag"
	"fmt"
	"os"
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

	o := patchutil.Options{
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
		return true, unpackCmd(o, args...)
	case "pack":
		check(false, 4)
		return true, packCmd(o, args...)
	case "packwithidx":
		check(false, 4)
		return true, packWithIdxCmd(o, args...)
	case "unpackfromfile":
		check(false, 3)
		return true, unpackFromFileCmd(o, args...)
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
