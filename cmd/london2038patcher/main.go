package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patcher"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/strutil"
	"github.com/ricochhet/london2038patcher/pkg/winutil"
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() {
	logutil.Infof(logutil.Get(), "patcher-%s\n", gitHash)
	logutil.Infof(logutil.Get(), "Build date: %s\n", buildDate)
	logutil.Infof(logutil.Get(), "Build on: %s\n", buildOn)
	os.Exit(0)
}

func main() {
	logutil.LogTime.Store(true)
	logutil.MaxProcNameLength.Store(0)
	logutil.Set(logutil.NewLogger("patcher", 0))
	logutil.SetDebug(flags.Debug)
	_ = cmdutil.QuickEdit(flags.QuickEdit)

	cmd, err := commands()
	if err != nil {
		logutil.Errorf(logutil.Get(), "Error running command: %v\n", err)
	}

	if cmd {
		return
	}

	p := patcher.NewContext()
	p.Set(&patcher.Patcher{
		HTTPClient:    *dlutil.NewHTTPClient(time.Duration(flags.Timeout)),
		ChecksumURL:   flags.ChecksumURL,
		PatchURL:      flags.PatchURL,
		ChecksumFile:  flags.ChecksumFile,
		HellgateCUKey: "",
		HellgateKey:   "",
		UsePatchDir:   flags.PatchDir,
		PatchDir:      "",
	})

	if err := downloadCmd(p); err != nil {
		logutil.Errorf(logutil.Get(), "%w\n", err)
	}
}

// commands handles the specified command flags.
func commands() (bool, error) {
	args := flag.Args()
	if len(args) == 0 {
		return false, nil
	}

	cmd := strings.ToLower(args[0])
	rest := args[1:]

	lr := patchutil.NewDefaultLocaleRegistry()

	lf, err := patchutil.NewLocaleFilter(lr, strutil.ToSlice(flags.Locales, ","))
	if err != nil {
		return true, err
	}

	o := patchutil.Options{
		Registry: lr,
		Filter:   lf,
		IdxOptions: &patchutil.IdxOptions{
			CRC32: flags.CRC32,
		},
		Archs: strutil.ToSlice(flags.Archs, ","),
	}

	switch cmd {
	case "decodeidx":
		cmds.Check(2)
		return true, decodeCmd(rest...)
	case "encodeidx":
		cmds.Check(2)
		return true, encodeCmd(rest...)
	case "unpack":
		cmds.Check(3)
		return true, unpackCmd(o, rest...)
	case "pack":
		cmds.Check(3)
		return true, packCmd(o, rest...)
	case "packwithidx":
		cmds.Check(3)
		return true, packWithIdxCmd(o, rest...)
	case "unpackfromfile":
		cmds.Check(2)
		return true, unpackFromFileCmd(o, rest...)
	case "regedit":
		cmds.Check(2)
		cmdutil.Supports("windows")

		return true, regeditCmd(rest...)
	case "help", "h":
		cmds.Usage()
	case "version", "v":
		version()
	default:
		cmds.Usage()
	}

	if winutil.IsAdmin() {
		cmdutil.Pause()
	}

	return false, nil
}
