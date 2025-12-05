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
func (m *PatcherCtx) Get() *Patcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.patcher
}

// Set sets the patcher.
func (m *PatcherCtx) Set(patcher *Patcher) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.patcher = patcher
}

// CopyFrom sets all patcher to the target.
func (m *PatcherCtx) CopyFrom(target *PatcherCtx) {
	m.Set(target.Get())
}
