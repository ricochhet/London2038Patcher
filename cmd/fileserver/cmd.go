package main

import (
	"path/filepath"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/server"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// dumpCmd command.
func dumpCmd(a ...string) error {
	return Embed().Dump(a[0], "", func(f embedutil.File, b []byte) error {
		logutil.Infof(logutil.Get(), "Writing: %s (%d bytes)\n", f.Path, f.Info.Size())
		return fsutil.Write(filepath.Join("dump", f.Path), b)
	})
}

// listCmd command.
func listCmd(a ...string) error {
	return Embed().List(a[0], func(files []embedutil.File) error {
		for _, f := range files {
			logutil.Infof(logutil.Get(), "%s (%d bytes)\n", f.Path, f.Info.Size())
		}

		return nil
	})
}

// serverCmd command.
func serverCmd(s *server.Context) error {
	if err := s.StartServer(); err != nil {
		logutil.Errorf(logutil.Get(), "Error starting server: %v\n", err)
		return errutil.WithFrame(err)
	}

	return nil
}
