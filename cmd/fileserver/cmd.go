package main

import (
	"path/filepath"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/server"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

// dumpCmd command.
func dumpCmd(a ...string) error {
	return timeutil.Timer(func() error {
		err := Embed().Dump(a[0], "", func(f embedutil.File, b []byte) error {
			logutil.Infof(logutil.Get(), "Writing: %s (%d bytes)\n", f.Path, f.Info.Size())
			return fsutil.Write(filepath.Join("dump", f.Path), b)
		})
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error dumping embedded files: %v\n", err)
		}

		return err
	}, "Dump", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// listCmd command.
func listCmd(a ...string) error {
	return timeutil.Timer(func() error {
		err := Embed().List(a[0], func(files []embedutil.File) error {
			for _, f := range files {
				logutil.Infof(logutil.Get(), "%s (%d bytes)\n", f.Path, f.Info.Size())
			}

			return nil
		})
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error listing embedded files: %v\n", err)
		}

		return err
	}, "List", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}

// serverCmd command.
func serverCmd(s *server.Context) error {
	return timeutil.Timer(func() error {
		err := s.StartServer()
		if err != nil {
			logutil.Errorf(logutil.Get(), "Error starting server: %v\n", err)
		}

		return err
	}, "List", func(_, elapsed string) {
		logutil.Infof(logutil.Get(), "Took %s\n", elapsed)
	})
}
