package server

import (
	"os"
	"strings"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/pkg/cryptoutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// maybeBase64 checks if name contains a prefix associated with embedded files. If it does, return the embedded file as base64.
func maybeBase64(fs *embedutil.EmbeddedFileSystem, name string) ([]byte, error) {
	if after, ok := strings.CutPrefix(name, "asset:"); ok {
		return maybeRead(fs, after), nil
	}

	b, err := cryptoutil.DecodeB64(name)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return b, nil
}

// maybeRead reads the specified name from the embedded filesystem. If it cannot be read, the program will exit.
func maybeRead(fs *embedutil.EmbeddedFileSystem, name string) []byte {
	b, err := fs.Read(name)
	if err != nil {
		logutil.Errorf(logutil.Get(), "Error reading from embedded filesystem: %v\n", err)
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
						Route:  "/",
						Name:   "index.html",
						Base64: cryptoutil.EncodeB64(maybeRead(c.FS, "index.html")),
					},
					{
						Route:  "/404.html",
						Name:   "404.html",
						Base64: cryptoutil.EncodeB64(maybeRead(c.FS, "404.html")),
					},
					{
						Route:  "/base.css",
						Name:   "base.css",
						Base64: cryptoutil.EncodeB64(maybeRead(c.FS, "base.css")),
					},
				},
			},
		},
	}
}
