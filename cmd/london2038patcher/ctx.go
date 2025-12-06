package main

import "sync"

type PatcherCtx struct {
	mu      sync.Mutex
	patcher *Patcher
}

// NewPatcherCtx creates an empty PatcherCtx.
func NewPatcherCtx() *PatcherCtx {
	return &PatcherCtx{}
}

// Get returns the patcher.
func (p *PatcherCtx) Get() *Patcher {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.patcher
}

// Set sets the patcher.
func (p *PatcherCtx) Set(patcher *Patcher) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.patcher = patcher
}

// CopyFrom sets all patcher to the target.
func (p *PatcherCtx) CopyFrom(target *PatcherCtx) {
	p.Set(target.Get())
}
