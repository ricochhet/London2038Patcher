package serverutil

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
)

// WithLogging is a middleware that logs the method and URL path for the handler.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ServeFileHandler creates a Handler for http.ServeFile.
func ServeFileHandler(info configutil.Info, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setHeaders(nil, info, w)
		http.ServeFile(w, r, name)
	})
}

// ServeContentHandler creates a Handler for http.ServeContent.
func ServeContentHandler(info configutil.Info, name string, data []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setHeaders(data, info, w)
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(data))
	})
}

// setHeaders sets the headers for the http.ResponseWriter.
func setHeaders(data []byte, info configutil.Info, w http.ResponseWriter) {
	set := true

	if info.StatusCode != 0 {
		w.WriteHeader(info.StatusCode)
	}

	for key, value := range info.Headers {
		if key == "Content-Type" {
			set = false
		}

		w.Header().Set(key, value)
	}

	if set && len(data) != 0 {
		w.Header().Set("Content-Type", http.DetectContentType(data))
	}
}
