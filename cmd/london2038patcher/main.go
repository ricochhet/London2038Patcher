package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
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

	_ = patcher.download()
}

// commands handles the specified command flags.
func commands() (bool, error) {
	if flag.NArg() == 0 {
		return false, nil
	}

	cmd := strings.ToLower(flag.Args()[0])
	args := flag.Args()[1:]

	d := patchutil.Dat{
		LocaleMap: patchutil.NewDefaultLocales(),
		Locales:   toSlice(Flag.Locales, ","),
		Archs:     toSlice(Flag.Archs, ","),
	}

	switch cmd {
	case "decodeidx":
		check(false, 3)
		return true, decode(args...)
	case "encodeidx":
		check(false, 3)
		return true, encode(args...)
	case "unpack":
		check(false, 4)
		return true, unpack(d, args...)
	case "pack":
		check(false, 4)
		return true, pack(d, args...)
	case "packwithidx":
		check(false, 4)
		return true, packWithIdx(d, args...)
	case "unpackfromfile":
		check(false, 3)
		return true, unpackFromFile(d, args...)
	case "regedit":
		check(true, 3)
		return true, regedit(args...)
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

// decode command.
func decode(a ...string) error {
	return timeutil.Timer(func() error {
		_, dErr := patchutil.DecodeFile(a[0], a[1])
		if dErr != nil {
			fmt.Fprintf(os.Stderr, "Error decoding index file: %v\n", dErr)
		}

		return dErr
	}, "Decode", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// encode command.
func encode(a ...string) error {
	return timeutil.Timer(func() error {
		eErr := patchutil.EncodeFile(a[0], a[1])
		if eErr != nil {
			fmt.Fprintf(os.Stderr, "Error encoding index file: %v\n", eErr)
		}

		return eErr
	}, "Encode", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// unpack command.
func unpack(d patchutil.Dat, a ...string) error {
	return timeutil.Timer(func() error {
		uErr := d.Unpack(a[0], a[1], a[2])
		if uErr != nil {
			fmt.Fprintf(os.Stderr, "Error unpacking patch: %v\n", uErr)
		}

		return uErr
	}, "Unpack", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// pack command.
func pack(d patchutil.Dat, a ...string) error {
	return timeutil.Timer(func() error {
		pErr := d.Pack(a[0], a[1], a[2])
		if pErr != nil {
			fmt.Fprintf(os.Stderr, "Error packing patch: %v\n", pErr)
		}

		return pErr
	}, "Pack", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// packWithIdx command.
func packWithIdx(d patchutil.Dat, a ...string) error {
	return timeutil.Timer(func() error {
		pErr := d.PackWithIndex(a[0], a[1], a[2])
		if pErr != nil {
			fmt.Fprintf(os.Stderr, "Error packing patch: %v\n", pErr)
		}

		return pErr
	}, "PackWithIndex", func(_, elapsed string) {
		fmt.Fprintf(os.Stdout, "Took %s\n", elapsed)
	})
}

// unpackFromFile command.
func unpackFromFile(d patchutil.Dat, a ...string) error {
	return timeutil.Timer(func() error {
		uErr := d.UnpackFromFile(a[0], a[1])
		if uErr != nil {
			fmt.Fprintf(os.Stderr, "Error unpacking patch: %v\n", uErr)
		}

		return uErr
	}, "UnpackFromFile", func(_, elapsed string) {
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
