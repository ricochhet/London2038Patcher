package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/serverutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/jsonutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

type Context struct {
	ConfigFile string
	TLS        *configutil.TLS
	FS         *embedutil.EmbeddedFileSystem

	servers []*http.Server
}

// NewServer returns a new Server type with assets preloaded.
func NewServer(configFile string, tls *configutil.TLS, fs *embedutil.EmbeddedFileSystem) *Context {
	s := &Context{}
	if configFile != "" {
		s.ConfigFile = configFile
	}

	s.TLS = tls
	s.FS = fs

	return s
}

// StartServer starts an HTTP server with the specified server configuration.
func (c *Context) StartServer() error {
	config, err := c.maybeReadConfig(c.ConfigFile)
	if err != nil {
		return errutil.New("c.maybeReadConfig", err)
	}

	c.maybeTLS(config)

	for _, cfg := range config.Servers {
		ctx := serverutil.NewHTTPServerCtx()

		maxAge := 300
		if cfg.MaxAge != 0 {
			maxAge = cfg.MaxAge
		}

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(serverutil.WithLogging)
		r.Use(httprate.LimitByIP(50, time.Minute))
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           maxAge,
		}))

		r.NotFound(c.NotFoundHandler)

		ctx.Set(&serverutil.HTTPServer{
			Router:   r,
			TLS:      c.TLS,
			Timeouts: &cfg.Timeouts,
		})

		if err := c.startServer(ctx, &cfg); err != nil {
			return errutil.New("c.startServer", err)
		}
	}

	c.shutdown()

	return nil
}

// shutdown handles shutdown of all servers.
func (c *Context) shutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, srv := range c.servers {
		if err := srv.Shutdown(ctx); err != nil {
			logutil.Errorf(logutil.Get(), "Error shutting down server: %v\n", err)
		}
	}
}

// maybeTLS sets TLS based on whether flags are based, or if relevant config options are used.
func (c *Context) maybeTLS(cfg *configutil.Config) {
	if c.TLS.CertFile == "" || c.TLS.KeyFile == "" { // default flags
		c.TLS.Enabled = false
	}

	if fsutil.Exists(c.TLS.CertFile) && fsutil.Exists(c.TLS.KeyFile) { // flags
		c.TLS.Enabled = true
		return
	}

	if fsutil.Exists(cfg.TLS.CertFile) && fsutil.Exists(cfg.TLS.KeyFile) { // config
		c.TLS = &cfg.TLS
	}
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
			logutil.Errorf(logutil.Get(), "Error reading server config: %v\n", err)
		}

		return config, err
	case !exists && path != "":
		return nil, fmt.Errorf("path specified but does not exist: %s", path)
	default:
		logutil.Infof(logutil.Get(), "Starting with default server config\n")
		return c.newDefaultConfig(), nil
	}
}

// startServer starts an HTTP server with the specified server configuration.
func (c *Context) startServer(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) error {
	if err := serveFileHandler(ctx, cfg); err != nil {
		return errutil.New("serveFileHandler", err)
	}

	if err := c.serveContentHandler(ctx, cfg); err != nil {
		return errutil.New("c.serveContentHandler", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := ctx.ListenAndServe(addr)

	c.servers = append(c.servers, srv)

	return nil
}

// serveContentHandler handles the ServeFileHandler for each file entry.
func serveFileHandler(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) error {
	for _, f := range cfg.FileEntries {
		info, err := os.Stat(f.Path)
		if err != nil {
			return errutil.WithFramef("invalid path %s: %w", f.Path, err)
		}

		if info.IsDir() {
			if err := matchPattern(f, ctx, cfg); err != nil {
				return errutil.New("matchPattern", err)
			}
		} else {
			if err := matchFile(f, ctx, cfg); err != nil {
				return errutil.New("matchFile", err)
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
			return errutil.WithFrame(err)
		}

		if info.IsDir() {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return errutil.WithFramef("invalid path %s: %w", path, err)
		}

		rel, err := filepath.Rel(f.Path, path)
		if err != nil {
			return errutil.WithFramef("cannot get relative path for %s: %w", path, err)
		}

		route := filepath.ToSlash(filepath.Join(f.Route, rel))

		logutil.Infof(logutil.Get(), "Port %d: %s -> %s\n", cfg.Port, route, abs)
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
		return errutil.WithFramef("invalid path %s: %w", f.Path, err)
	}

	logutil.Infof(logutil.Get(), "Port %d: %s -> %s\n", cfg.Port, f.Route, abs)
	ctx.Handle(f.Route, serverutil.ServeFileHandler(f.Info, abs))

	return nil
}

// serveContentHandler handles the ServeContentHandler for each content entry.
func (c *Context) serveContentHandler(ctx *serverutil.HTTPServerCtx, cfg *configutil.Server) error {
	for _, f := range cfg.ContentEntries {
		logutil.Infof(
			logutil.Get(),
			"Port %d: %s -> %s (%d)\n",
			cfg.Port,
			f.Route,
			f.Name,
			len(f.Base64),
		)

		b, err := maybeBase64(c.FS, f.Base64)
		if err != nil {
			return errutil.WithFrame(err)
		}

		ctx.Handle(f.Route, serverutil.ServeContentHandler(f.Info, f.Name, b))
	}

	return nil
}
