package serverutil

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

type headerWriter struct {
	http.ResponseWriter

	statusCode int
	allowed    map[string]struct{}
}

// WriteHeader strips non-allowed headers when writing to the actual header.
func (h *headerWriter) WriteHeader(code int) {
	hdr := h.Header()
	for key := range hdr {
		if _, ok := h.allowed[http.CanonicalHeaderKey(key)]; !ok {
			hdr.Del(key)
		}
	}

	if h.statusCode != 0 {
		code = h.statusCode
	}

	h.ResponseWriter.WriteHeader(code)
}

// newHeaderWriter sets the allowed headers, returning a new headerWriter.
func newHeaderWriter(w http.ResponseWriter, data []byte, info configutil.Info) *headerWriter {
	allowed := make(map[string]struct{})
	setContentType := false

	for key, value := range info.Headers {
		canonical := http.CanonicalHeaderKey(key)
		w.Header().Set(canonical, value)
		allowed[canonical] = struct{}{}

		if canonical == "Content-Type" {
			setContentType = true
		}
	}

	if !setContentType {
		if len(data) != 0 {
			ct := http.DetectContentType(data)
			w.Header().Set("Content-Type", ct)
		}

		allowed["Content-Type"] = struct{}{}
	}

	return &headerWriter{
		ResponseWriter: w,
		statusCode:     info.StatusCode,
		allowed:        allowed,
	}
}

// WithLogging is a middleware that logs the method and URL path for the handler.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logutil.Infof(logutil.Get(), "%s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ServeFileHandler creates a Handler for http.ServeFile.
func ServeFileHandler(info configutil.Info, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(newHeaderWriter(w, nil, info), r, name)
	})
}

// ServeContentHandler creates a Handler for http.ServeContent.
func ServeContentHandler(info configutil.Info, name string, data []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(
			newHeaderWriter(w, data, info),
			r,
			name,
			time.Now(),
			bytes.NewReader(data),
		)
	})
}
