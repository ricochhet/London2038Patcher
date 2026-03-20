package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/browse"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/hostsutil"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/serverutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/jsonutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

type Context struct {
	ConfigFile string
	Hosts      bool
	TLS        *configutil.TLS
	FS         *embedutil.EmbeddedFileSystem

	servers []*http.Server
}

// NewServer returns a new Server type with assets preloaded.
func NewServer(
	configFile string,
	hosts bool,
	tls *configutil.TLS,
	fs *embedutil.EmbeddedFileSystem,
) *Context {
	s := &Context{}
	if configFile != "" {
		s.ConfigFile = configFile
	}

	s.Hosts = hosts
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

	if err := c.addHosts(config); err != nil {
		return errutil.New("c.addHosts", err)
	}

	c.maybeTLS(config)

	for _, cfg := range config.Servers {
		ctx := serverutil.NewContext()

		maxAge := 300
		if cfg.MaxAge != 0 {
			maxAge = cfg.MaxAge
		}

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(serverutil.WithLogging)
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           maxAge,
		}))

		r.NotFound(c.NotFoundHandler)

		ctx.SetLocked(&serverutil.HTTPServer{
			Router:   r,
			TLS:      c.TLS,
			Timeouts: &cfg.Timeouts,
		})

		if cfg.FormAuth.Username != "" && cfg.FormAuth.Password != "" {
			secret := resolveFormAuthSecret(cfg.FormAuth.Secret)
			r.Use(withFormAuth(secret, cfg.FormAuth.PublicPrefixes))
			c.registerAuthRoutes(ctx.Handle, cfg.FormAuth.Username, cfg.FormAuth.Password, secret)
			logutil.Infof(
				logutil.Get(),
				"Port %d: form auth enabled for user %q\n",
				cfg.Port,
				cfg.FormAuth.Username,
			)
		} else if cfg.BasicAuth.Username != "" && cfg.BasicAuth.Password != "" {
			r.Use(withBasicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password))
			logutil.Infof(
				logutil.Get(),
				"Port %d: basic auth enabled for user %q\n",
				cfg.Port,
				cfg.BasicAuth.Username,
			)
		}

		if err := c.startServer(ctx, &cfg); err != nil {
			return errutil.New("c.startServer", err)
		}
	}

	if err := c.removeHosts(config); err != nil {
		return errutil.New("c.removeHosts", err)
	}

	c.shutdown()

	return nil
}

// withBasicAuth returns a chi middleware that enforces HTTP Basic Authentication.
func withBasicAuth(user, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || u != user || p != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="fileserver"`)
				errutil.HTTPUnauthorized(w)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// wrapBasicAuth wraps a single handler with Basic Auth when credentials are non-empty.
func wrapBasicAuth(auth configutil.BasicAuth, h http.Handler) http.Handler {
	if auth.Username == "" || auth.Password == "" {
		return h
	}

	return withBasicAuth(auth.Username, auth.Password)(h)
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

// addHosts adds the specified hosts from the configuration.
func (c *Context) addHosts(cfg *configutil.Config) error {
	if !c.isHostsValid(cfg) {
		return nil
	}

	hf, err := hostsutil.NewHosts()
	if err != nil {
		return errutil.New("hostsutil.NewHosts", err)
	}

	return hostsutil.Add(hf, cfg.Hosts)
}

// removeHosts removes the specified hosts from the configuration.
func (c *Context) removeHosts(cfg *configutil.Config) error {
	if !c.isHostsValid(cfg) {
		return nil
	}

	hf, err := hostsutil.NewHosts()
	if err != nil {
		return errutil.New("hostsutil.NewHosts", err)
	}

	return hostsutil.Remove(hf, cfg.Hosts)
}

// isHostsValid returns if the hosts state is valid.
func (c *Context) isHostsValid(cfg *configutil.Config) bool {
	return c.Hosts && cfg.Hosts != nil && len(cfg.Hosts) != 0
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
func (c *Context) startServer(ctx *serverutil.Context, cfg *configutil.Server) error {
	browseRateLimit := 500
	if cfg.BrowseRateLimit != 0 {
		browseRateLimit = cfg.BrowseRateLimit
	}

	fileRateLimit := 100
	if cfg.FileRateLimit != 0 {
		fileRateLimit = cfg.FileRateLimit
	}

	browseLimit := httprate.LimitByIP(browseRateLimit, time.Minute)
	fileLimit := httprate.LimitByIP(fileRateLimit, time.Minute)

	if err := c.serveFileHandler(ctx, cfg, browseLimit, fileLimit); err != nil {
		return errutil.New("c.serveFileHandler", err)
	}

	if err := c.serveContentHandler(ctx, cfg, browseLimit); err != nil {
		return errutil.New("c.serveContentHandler", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := ctx.ListenAndServe(addr)

	c.servers = append(c.servers, srv)

	return nil
}

// serveFileHandler handles the ServeFileHandler for each file entry.
func (c *Context) serveFileHandler(
	ctx *serverutil.Context,
	cfg *configutil.Server,
	browseLimit, fileLimit func(http.Handler) http.Handler,
) error {
	for _, f := range cfg.FileEntries {
		info, err := os.Stat(f.Path)
		if err != nil {
			return errutil.WithFramef("invalid path %s: %w", f.Path, err)
		}

		if info.IsDir() && f.Browse != "" {
			route := strings.TrimSuffix(f.Browse, "/")
			handler := browse.Handler(
				c.FS, f.Path, route, cfg.Hidden,
				cfg,
			)
			handler = browseLimit(wrapBasicAuth(f.BasicAuth, handler))

			logutil.Infof(logutil.Get(), "Port %d: %s/** -> %s (browse)\n", cfg.Port, route, f.Path)

			ctx.Handle(route, handler)
			ctx.Handle(route+"/*", handler)
		}

		if info.IsDir() {
			if err := matchPattern(f, ctx, cfg, fileLimit); err != nil {
				return errutil.New("matchPattern", err)
			}
		} else {
			if err := matchFile(f, ctx, cfg, fileLimit); err != nil {
				return errutil.New("matchFile", err)
			}
		}
	}

	return nil
}

// serveContentHandler handles the ServeContentHandler for each content entry.
func (c *Context) serveContentHandler(
	ctx *serverutil.Context,
	cfg *configutil.Server,
	limit func(http.Handler) http.Handler,
) error {
	for _, f := range cfg.ContentEntries {
		logutil.Infof(
			logutil.Get(),
			"Port %d: %s -> %s (%d)\n",
			cfg.Port,
			f.Route,
			f.Name,
			len(f.Base64),
		)

		b, err := embedutil.MaybeBase64(c.FS, f.Base64)
		if err != nil {
			return errutil.WithFrame(err)
		}

		ctx.Handle(f.Route, limit(serverutil.ServeContentHandler(f.Info, f.Name, b)))
	}

	return nil
}

// matchPattern handles file paths that contain glob information.
func matchPattern(
	f configutil.FileEntry,
	ctx *serverutil.Context,
	cfg *configutil.Server,
	fileLimit func(http.Handler) http.Handler,
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

		handler := fileLimit(wrapBasicAuth(f.BasicAuth, serverutil.ServeFileHandler(f.Info, abs)))
		ctx.Handle(route, handler)

		return nil
	})
}

// matchFile handles absolute file paths.
func matchFile(
	f configutil.FileEntry,
	ctx *serverutil.Context,
	cfg *configutil.Server,
	fileLimit func(http.Handler) http.Handler,
) error {
	abs, err := filepath.Abs(f.Path)
	if err != nil {
		return errutil.WithFramef("invalid path %s: %w", f.Path, err)
	}

	logutil.Infof(logutil.Get(), "Port %d: %s -> %s\n", cfg.Port, f.Route, abs)

	handler := fileLimit(wrapBasicAuth(f.BasicAuth, serverutil.ServeFileHandler(f.Info, abs)))
	ctx.Handle(f.Route, handler)

	return nil
}
