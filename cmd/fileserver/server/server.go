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

type Context struct {
	ConfigFile string
	FS         *embedutil.EmbeddedFileSystem
}

// NewServer returns a new Server type with assets preloaded.
func NewServer(configFile string, fs *embedutil.EmbeddedFileSystem) *Context {
	s := &Context{}
	if configFile != "" {
		s.ConfigFile = configFile
	}

	s.FS = fs

	return s
}

// StartServer starts an HTTP server with the specified server configuration.
func (c *Context) StartServer() error {
	config, err := c.maybeReadConfig(c.ConfigFile)
	if err != nil {
		return err
	}

	for _, cfg := range config.Servers {
		ctx := serverutil.NewHTTPServerCtx()

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(serverutil.WithLogging)

		r.NotFound(c.NotFoundHandler)

		ctx.Set(&serverutil.HTTPServer{
			Router: r,
			TLS:    *serverutil.NewTLS(),
		})

		if err := startServer(ctx, &cfg); err != nil {
			return err
		}
	}

	return nil
}

// maybeReadConfig reads the file path if it exists, otherwise returning a default config.
func (c *Context) maybeReadConfig(path string) (*configutil.Config, error) {
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
		return c.newDefaultConfig(), nil
	}
}

// startServer starts an HTTP server with the specified server configuration.
func startServer(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) error {
	if err := serveFileHandler(ctx, cfg); err != nil {
		return err
	}

	serveContentHandler(ctx, cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	go ctx.ListenAndServe(addr)

	return nil
}

// serveContentHandler handles the ServeFileHandler for each file entry.
func serveFileHandler(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) error {
	for _, f := range cfg.FileEntries {
		info, err := os.Stat(f.Path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", f.Path, err)
		}

		if info.IsDir() {
			if err := matchPattern(f, ctx, cfg); err != nil {
				return err
			}
		} else {
			if err := matchFile(f, ctx, cfg); err != nil {
				return err
			}
		}
	}

	return nil
}

// matchPattern handles file paths that contain glob information.
func matchPattern(
	f configutil.FileEntry,
	ctx *serverutil.HTTPServerCtx,
	cfg *configutil.Server,
) error {
	return filepath.Walk(f.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		rel, err := filepath.Rel(f.Path, path)
		if err != nil {
			return fmt.Errorf("cannot get relative path for %s: %w", path, err)
		}

		route := filepath.ToSlash(filepath.Join(f.Route, rel))

		fmt.Fprintf(os.Stdout, "Port %d: %s -> %s\n", cfg.Port, route, abs)
		ctx.Handle(route, serverutil.ServeFileHandler(f.Info, abs))

		return nil
	})
}

// matchFile handles absolute file paths.
func matchFile(
	f configutil.FileEntry,
	ctx *serverutil.HTTPServerCtx,
	cfg *configutil.Server,
) error {
	abs, err := filepath.Abs(f.Path)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", f.Path, err)
	}

	fmt.Fprintf(os.Stdout, "Port %d: %s -> %s\n", cfg.Port, f.Route, abs)
	ctx.Handle(f.Route, serverutil.ServeFileHandler(f.Info, abs))

	return nil
}

// serveContentHandler handles the ServeContentHandler for each content entry.
func serveContentHandler(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) {
	for _, f := range cfg.ContentEntries {
		fmt.Fprintf(os.Stdout, "Port %d: %s -> %s (%d)\n", cfg.Port, f.Route, f.Name, len(f.Bytes))

		ctx.Handle(f.Route, serverutil.ServeContentHandler(f.Info, f.Name, f.Bytes))
	}
}
