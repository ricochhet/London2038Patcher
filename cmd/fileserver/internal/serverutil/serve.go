package serverutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// listenAndServe creates an HTTP server at the specified address.
func (h *HTTPServerCtx) ListenAndServe(addr string) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: h.httpServer.Router,
	}

	logutil.Infof(logutil.Get(), "Server listening on %s\n", addr)

	var err error

	if h.httpServer.TLS.Enabled {
		fmt.Fprintf(
			os.Stdout,
			"Server starting with tls: %s (cert) and %s (key)\n",
			h.httpServer.TLS.CertFile, h.httpServer.TLS.KeyFile,
		)
		err = server.ListenAndServeTLS(h.httpServer.TLS.CertFile, h.httpServer.TLS.KeyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logutil.Infof(logutil.Get(), "Server %s failed: %v\n", strings.TrimPrefix(addr, ":"), err)
	}

	return server
}
