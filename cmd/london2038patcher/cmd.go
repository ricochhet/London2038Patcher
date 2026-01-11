package main

import (
	"fmt"
	"os"

	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/regutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

// downloadCmd command.
func (p *PatcherCtx) downloadCmd() error {
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

// decodeCmd command.
func decodeCmd(a ...string) error {
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

// encodeCmd command.
func encodeCmd(a ...string) error {
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

// unpackCmd command.
func unpackCmd(d patchutil.Options, a ...string) error {
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

// packCmd command.
func packCmd(d patchutil.Options, a ...string) error {
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

// packWithIdxCmd command.
func packWithIdxCmd(d patchutil.Options, a ...string) error {
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

// unpackFromFileCmd command.
func unpackFromFileCmd(d patchutil.Options, a ...string) error {
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

// regeditCmd command.
func regeditCmd(a ...string) error {
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
