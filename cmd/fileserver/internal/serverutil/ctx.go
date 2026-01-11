package serverutil

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	Router chi.Router

	TLS TLS
}

type TLS struct {
	Cert string
	Key  string
}

type HTTPServerCtx struct {
	mu         sync.Mutex
	httpServer *HTTPServer
}

// NewTLS creates an empty TLS.
func NewTLS() *TLS {
	return &TLS{}
}

// NewHTTPServerCtx creates an empty PatcherCtx.
func NewHTTPServerCtx() *HTTPServerCtx {
	return &HTTPServerCtx{}
}

// Get returns the patcher.
func (h *HTTPServerCtx) Get() *HTTPServer {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.httpServer
}

// Set sets the patcher.
func (h *HTTPServerCtx) Set(server *HTTPServer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.httpServer = server
}

// CopyFrom sets all patcher to the target.
func (h *HTTPServerCtx) CopyFrom(target *HTTPServerCtx) {
	h.Set(target.Get())
}

func (h *HTTPServerCtx) Handle(pattern string, handler http.Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.httpServer.Router.Handle(pattern, handler)
}
