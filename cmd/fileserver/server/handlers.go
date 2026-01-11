package server

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// NotFoundHandler is a middleware for 404 not found.
func (s *Server) NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(maybeRead(s.FS, "404.html"))
}

// SPANotFound returns a SPA-style fallback HandlerFunc.
func SPANotFound(name string, b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(filepath.Base(r.URL.Path), ".") {
			http.NotFound(w, r)
			return
		}

		http.ServeContent(
			w,
			r,
			name,
			time.Now(),
			bytes.NewReader(b),
		)
	}
}
