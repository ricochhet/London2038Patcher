package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/server"
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() string {
	return fmt.Sprintf(
		"fileserver\n\tBuild Date: %s\n\tGit Hash: %s\n\tBuilt On: %s\n",
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

	_, err := commands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
	}
}

// commands handles the specified command flags.
func commands() (bool, error) {
	var cmd string

	if flag.NArg() != 0 {
		cmd = strings.ToLower(flag.Args()[0])
	}

	switch cmd {
	case "help", "h":
		usage()
	default:
		d := server.NewServer(cmd, Embed())
		return true, serverCmd(d)
	}

	return false, nil
}
