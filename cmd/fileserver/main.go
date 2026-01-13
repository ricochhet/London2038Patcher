package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/server"
	"github.com/ricochhet/london2038patcher/pkg/cmdutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/winutil"
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() {
	logutil.Infof(logutil.Get(), "fileserver-%s\n", gitHash)
	logutil.Infof(logutil.Get(), "Build date: %s\n", buildDate)
	logutil.Infof(logutil.Get(), "Build on: %s\n", buildOn)
	os.Exit(0)
}

func main() {
	logutil.LogTime.Store(true)
	logutil.MaxProcNameLength.Store(0)
	logutil.Set(logutil.NewLogger("fileserver", 0))
	logutil.SetDebug(flags.Debug)
	_ = cmdutil.QuickEdit(flags.QuickEdit)

	cmd, err := commands()
	if err != nil {
		logutil.Errorf(logutil.Get(), "Error running command: %v\n", err)
	}

	if cmd {
		return
	}

	s := server.NewServer(flags.ConfigFile, &configutil.TLS{
		Enabled:  true,
		CertFile: flags.CertFile,
		KeyFile:  flags.KeyFile,
	}, Embed())
	if err := serverCmd(s); err != nil {
		logutil.Errorf(logutil.Get(), "%v\n", err)
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

	switch cmd {
	case "dump", "d":
		cmds.Check(1)
		return true, dumpCmd(rest...)
	case "list", "l":
		cmds.Check(1)
		return true, listCmd(rest...)
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
