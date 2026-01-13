package main

import (
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patcher"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/patchutil"
	"github.com/ricochhet/london2038patcher/cmd/london2038patcher/internal/regutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

// downloadCmd command.
func downloadCmd(p *patcher.Context) error {
	return timeutil.Timer(func() error {
		err := p.Get().Download()
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error downloading files: %v\n", err)
		}

		return err
	}, "Download", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// decodeCmd command.
func decodeCmd(a ...string) error {
	return timeutil.Timer(func() error {
		_, err := patchutil.DecodeFile(a[0], a[1])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error decoding index file: %v\n", err)
		}

		return err
	}, "Decode", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// encodeCmd command.
func encodeCmd(a ...string) error {
	return timeutil.Timer(func() error {
		err := patchutil.EncodeFile(a[0], a[1])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error encoding index file: %v\n", err)
		}

		return err
	}, "Encode", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// unpackCmd command.
func unpackCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		err := o.Unpack(a[0], a[1], a[2])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error unpacking patch: %v\n", err)
		}

		return err
	}, "Unpack", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// packCmd command.
func packCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		err := o.Pack(a[0], a[1], a[2])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error packing patch: %v\n", err)
		}

		return err
	}, "Pack", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// packWithIdxCmd command.
func packWithIdxCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		err := o.PackWithIndex(a[0], a[1], a[2])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error packing patch: %v\n", err)
		}

		return err
	}, "PackWithIndex", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// unpackFromFileCmd command.
func unpackFromFileCmd(o patchutil.Options, a ...string) error {
	return timeutil.Timer(func() error {
		err := o.UnpackFromFile(a[0], a[1])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error unpacking patch: %v\n", err)
		}

		return err
	}, "UnpackFromFile", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// regeditCmd command.
func regeditCmd(a ...string) error {
	return timeutil.Timer(func() error {
		err := regutil.Regedit(a[0], a[1])
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error editing registry: %v\n", err)
		}

		return err
	}, "Regedit", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}
