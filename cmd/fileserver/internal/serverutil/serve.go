package serverutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// listenAndServe creates an HTTP server at the specified address.
func (h *HTTPServerCtx) ListenAndServe(addr string) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: h.httpServer.Router,
	}

	fmt.Fprintf(os.Stdout, "Server listening on %s\n", addr)

	var err error

	if h.httpServer.TLS.enabled() {
		fmt.Fprintf(
			os.Stdout,
			"Server starting with tls: %s (cert) and %s (key)\n",
			h.httpServer.TLS.Cert, h.httpServer.TLS.Key,
		)
		err = server.ListenAndServeTLS(h.httpServer.TLS.Cert, h.httpServer.TLS.Key)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stdout, "Server %s failed: %v\n", strings.TrimPrefix(addr, ":"), err)
	}

	return server
}

// enabled returns true if TLS is cert and key file is valid.
func (t *TLS) enabled() bool {
	return t.Cert != "" && t.Key != ""
}
