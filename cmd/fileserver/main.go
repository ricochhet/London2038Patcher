package main

import (
	"flag"
	"strings"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/server"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

func main() {
	logutil.LogTime.Store(true)
	logutil.MaxProcNameLength.Store(0)
	logutil.Set(logutil.NewLogger("fileserver", 0))

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

	s := server.NewServer(Flag.ConfigFile, &configutil.TLS{
		Enabled:  true,
		CertFile: Flag.CertFile,
		KeyFile:  Flag.KeyFile,
	}, Embed())
	_ = serverCmd(s)
}

// commands handles the specified command flags.
func commands() (bool, error) {
	var (
		cmd  string
		args []string
	)

	if flag.NArg() != 0 {
		cmd = strings.ToLower(flag.Args()[0])
	}

	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}

	switch cmd {
	case "help", "h":
		usage()
	case "dump", "d":
		check(1)
		return true, dumpCmd(args...)
	}

	return false, nil
}
