package server

import (
	"fmt"
	"os"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
)

// maybeRead reads the specified name from the embedded filesystem. If it cannot be read, the program will exit.
func maybeRead(fs *embedutil.EmbeddedFileSystem, name string) []byte {
	b, err := fs.Read(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from embedded filesystem: %v\n", err)
		os.Exit(1)
	}

	return b
}

// newDefaultConfig creates a default Config with the embedded index bytes.
func (c *Context) newDefaultConfig() *configutil.Config {
	return &configutil.Config{
		Servers: []configutil.Server{
			{
				Port: 8080,
				ContentEntries: []configutil.ContentEntry{
					{
						Route: "/",
						Name:  "index.html",
						Bytes: maybeRead(c.FS, "index.html"),
					},
					{
						Route: "/404.html",
						Name:  "404.html",
						Bytes: maybeRead(c.FS, "404.html"),
					},
					{
						Route: "/base.css",
						Name:  "base.css",
						Bytes: maybeRead(c.FS, "base.css"),
					},
				},
			},
		},
	}
}
