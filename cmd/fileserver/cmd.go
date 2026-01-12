package main

import (
	"encoding/base64"
	"flag"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/server"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// check handles checks for commands.
func check(v int) {
	if flag.NArg() < v {
		usage()
	}
}

// dumpCmd command.
func dumpCmd(a ...string) error {
	return Embed().Dump(a[0], "", func(f embedutil.File, b []byte) error {
		return fsutil.Write(
			filepath.Join("dump", f.Path+".base64"),
			[]byte(base64.StdEncoding.EncodeToString(b)),
		)
	})
}

// serverCmd command.
func serverCmd(s *server.Context) error {
	if err := s.StartServer(); err != nil {
		logutil.Errorf(logutil.Get(), "Error starting server: %v\n", err)
		return err
	}

	select {} // block
}
