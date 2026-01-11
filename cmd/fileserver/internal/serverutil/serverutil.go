package serverutil

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// listenAndServe creates an HTTP server at the specified address.
func (s *ServerCtx) ListenAndServe(addr string) {
	server := &http.Server{
		Addr:    addr,
		Handler: s.server,
	}

	fmt.Fprintf(os.Stdout, "Server listening on %s\n", addr)

	var err error

	if s.server.TLS.enabled() {
		fmt.Fprintf(
			os.Stdout,
			"Server starting with tls: %s (cert) and %s (key)\n",
			s.server.TLS.Cert, s.server.TLS.Key,
		)
		err = server.ListenAndServeTLS(s.server.TLS.Cert, s.server.TLS.Key)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stdout, "Server %s failed: %v\n", strings.TrimPrefix(addr, ":"), err)
	}
}

// WithLogging is a middleware that logs the method and URL path for the handler.
func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "%s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

// ServeFileHandler creates a HandlerFunc for http.ServeFile.
func ServeFileHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	}
}

// ServeContentHandler creates a HandlerFunc for http.ServeContent.
func ServeContentHandler(name string, b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(b))
	}
}

// enabled returns true if TLS is cert and key file is valid.
func (t *TLS) enabled() bool {
	return t.Cert != "" && t.Key != ""
}
