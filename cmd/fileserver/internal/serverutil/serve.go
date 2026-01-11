package serverutil

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// listenAndServe creates an HTTP server at the specified address.
func (s *ServerCtx) ListenAndServe(addr string) {
	server := &http.Server{
		Addr:    addr,
		Handler: s.server.Router,
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

// enabled returns true if TLS is cert and key file is valid.
func (t *TLS) enabled() bool {
	return t.Cert != "" && t.Key != ""
}
