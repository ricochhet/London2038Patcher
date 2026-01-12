package main

import (
	"flag"
	"strings"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/patcher"
	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/winutil"
)

func main() {
	logutil.LogTime.Store(true)
	logutil.MaxProcNameLength.Store(0)
	logutil.Set(logutil.NewLogger("patcher", 0))

	if Flag.Version {
		logutil.Info(logutil.Get(), version())
		return
	}

	cmd, err := commands()
	if err != nil {
		logutil.Errorf(logutil.Get(), "Error running command: %v\n", err)
	}

	if cmd {
		return
	}

	p := patcher.NewContext()
	p.Set(&patcher.Patcher{
		HTTPClient:    *dlutil.NewHTTPClient(time.Duration(Flag.Timeout)),
		ChecksumURL:   Flag.ChecksumURL,
		PatchURL:      Flag.PatchURL,
		ChecksumFile:  Flag.ChecksumFile,
		HellgateCUKey: "",
		HellgateKey:   "",
		UsePatchDir:   Flag.PatchDir,
		PatchDir:      "",
	})

	_ = downloadCmd(p)
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
