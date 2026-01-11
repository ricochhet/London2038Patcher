package serverutil

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

// WithLogging is a middleware that logs the method and URL path for the handler.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ServeFileHandler creates a Handler for http.ServeFile.
func ServeFileHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

// ServeContentHandler creates a Handler for http.ServeContent.
func ServeContentHandler(name string, b []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(b))
	})
}
