package patcher

import "sync"

type Context struct {
	mu      sync.Mutex
	patcher *Patcher
}

// NewContext creates an empty PatcherCtx.
func NewContext() *Context {
	return &Context{}
}

// Get returns the patcher.
func (p *Context) Get() *Patcher {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.patcher
}

// Set sets the patcher.
func (p *Context) Set(patcher *Patcher) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.patcher = patcher
}

// CopyFrom sets all patcher to the target.
func (p *Context) CopyFrom(target *Context) {
	p.Set(target.Get())
}
