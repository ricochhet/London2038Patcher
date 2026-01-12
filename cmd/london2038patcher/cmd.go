package main

import (
	"flag"
	"os"
	"runtime"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/regutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/patcher"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

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
		logutil.Errorf(logutil.Get(), "This command is unsupported on non-Windows machines.\n")
		os.Exit(1)
	}
}

// downloadCmd command.
func downloadCmd(p *patcher.Context) error {
	return timeutil.Timer(func() error {
		pErr := p.Get().Download()
		if pErr != nil {
			logutil.Errorf(logutil.Get(), "Error downloading files: %v\n", pErr)
		}

		return pErr
	}, "Download", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// decodeCmd command.
func decodeCmd(a ...string) error {
	return timeutil.Timer(func() error {
		_, dErr := patchutil.DecodeFile(a[0], a[1])
		if dErr != nil {
			logutil.Errorf(logutil.Get(), "Error decoding index file: %v\n", dErr)
		}

		return dErr
	}, "Decode", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// encodeCmd command.
func encodeCmd(a ...string) error {
	return timeutil.Timer(func() error {
		eErr := patchutil.EncodeFile(a[0], a[1])
		if eErr != nil {
			logutil.Errorf(logutil.Get(), "Error encoding index file: %v\n", eErr)
		}

		return eErr
	}, "Encode", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// unpackCmd command.
func unpackCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		uErr := o.Unpack(a[0], a[1], a[2])
		if uErr != nil {
			logutil.Errorf(logutil.Get(), "Error unpacking patch: %v\n", uErr)
		}

		return uErr
	}, "Unpack", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// packCmd command.
func packCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		pErr := o.Pack(a[0], a[1], a[2])
		if pErr != nil {
			logutil.Errorf(logutil.Get(), "Error packing patch: %v\n", pErr)
		}

		return pErr
	}, "Pack", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// packWithIdxCmd command.
func packWithIdxCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		pErr := o.PackWithIndex(a[0], a[1], a[2])
		if pErr != nil {
			logutil.Errorf(logutil.Get(), "Error packing patch: %v\n", pErr)
		}

		return pErr
	}, "PackWithIndex", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// unpackFromFileCmd command.
func unpackFromFileCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		uErr := o.UnpackFromFile(a[0], a[1])
		if uErr != nil {
			logutil.Errorf(logutil.Get(), "Error unpacking patch: %v\n", uErr)
		}

		return uErr
	}, "UnpackFromFile", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// regeditCmd command.
func regeditCmd(a ...string) error {
	return timeutil.Timer(func() error {
		rErr := regutil.Regedit(a[0], a[1])
		if rErr != nil {
			logutil.Errorf(logutil.Get(), "Error editing registry: %v\n", rErr)
		}

		return rErr
	}, "Regedit", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}
