package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/serverutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/jsonutil"
)

type Server struct {
	ConfigFile string
	FS         *embedutil.EmbeddedFileSystem
}

// NewServer returns a new Server type with assets preloaded.
func NewServer(configFile string, fs *embedutil.EmbeddedFileSystem) *Server {
	d := &Server{}
	if configFile != "" {
		d.ConfigFile = configFile
	}

	d.FS = fs

	return d
}

// StartServer starts an HTTP server with the specified server configuration.
func (s *Server) StartServer() error {
	ctx := serverutil.NewServerCtx()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(serverutil.WithLogging)

	r.NotFound(s.NotFoundHandler)

	ctx.Set(&serverutil.Server{
		Router: r,
		TLS:    *serverutil.NewTLS(),
	})

	config, err := s.maybeReadConfig(s.ConfigFile)
	if err != nil {
		return err
	}

	for _, server := range config.Servers {
		if err := startServer(ctx, &server); err != nil {
			return err
		}
	}

	return nil
}

// maybeReadConfig reads the file path if it exists, otherwise returning a default config.
func (s *Server) maybeReadConfig(path string) (*configutil.Config, error) {
	var (
		config *configutil.Config
		err    error
	)

	exists := fsutil.Exists(path)
	switch {
	case exists:
		config, err = jsonutil.ReadAndUnmarshal[configutil.Config](path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading server config: %v\n", err)
		}

		return config, err
	case !exists && path != "":
		return nil, fmt.Errorf("path specified but does not exist: %s", path)
	default:
		fmt.Fprintf(os.Stdout, "Starting with default server config\n")
		return s.newDefaultConfig(), nil
	}
}

// startServer starts an HTTP server with the specified server configuration.
func startServer(s *serverutil.ServerCtx, c *configutil.Server) error {
	for _, f := range c.Files {
		abs, err := filepath.Abs(f.Path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", f.Path, err)
		}

		fmt.Fprintf(os.Stdout, "Port %d: %s -> %s\n", c.Port, f.Route, abs)

		s.Handle(f.Route, serverutil.ServeFileHandler(abs))
	}

	for _, f := range c.Content {
		fmt.Fprintf(os.Stdout, "Port %d: %s -> %s (%d)\n", c.Port, f.Route, f.Name, len(f.Bytes))

		s.Handle(f.Route, serverutil.ServeContentHandler(f.Name, f.Bytes))
	}

	addr := fmt.Sprintf(":%d", c.Port)

	go s.ListenAndServe(addr)

	return nil
}
