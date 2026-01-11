package serverutil

import (
	"net/http"
	"sync"
)

type Server struct {
	*http.ServeMux

	TLS TLS
}

type TLS struct {
	Cert string
	Key  string
}

type ServerCtx struct {
	mu     sync.Mutex
	server *Server
}

// NewTLS creates an empty TLS.
func NewTLS() *TLS {
	return &TLS{}
}

// NewServerCtx creates an empty PatcherCtx.
func NewServerCtx() *ServerCtx {
	return &ServerCtx{}
}

// Get returns the patcher.
func (p *ServerCtx) Get() *Server {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.server
}

// Set sets the patcher.
func (p *ServerCtx) Set(server *Server) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.server = server
}

// CopyFrom sets all patcher to the target.
func (p *ServerCtx) CopyFrom(target *ServerCtx) {
	p.Set(target.Get())
}

func (p *ServerCtx) HandleFunc(pattern string, handler http.HandlerFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.server.HandleFunc(pattern, handler)
}
