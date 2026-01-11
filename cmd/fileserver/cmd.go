package main

import "github.com/ricochhet/london2038patcher/cmd/fileserver/server"

// serverCmd command.
func serverCmd(d *server.Server) error {
	if err := d.StartServer(); err != nil {
		return err
	}

	select {} // block
}
