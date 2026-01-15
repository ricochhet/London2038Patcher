package contextutil

import "github.com/sasha-s/go-deadlock"

type Context[T any] struct {
	Mutex deadlock.Mutex
	t     *T
}

// NewContext creates an empty Context.
func NewContext[T any]() *Context[T] {
	return &Context[T]{}
}

// GetLocked returns the patcher.
func (c *Context[T]) GetLocked() *T {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	return c.t
}

// Get is the same as GetLocked() but does not lock the mutex.
func (c *Context[T]) Get() *T {
	return c.t
}

// SetLocked sets the patcher.
func (c *Context[T]) SetLocked(t *T) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.t = t
}

// Set is the same as SetLocked() but does not lock the mutex.
func (c *Context[T]) Set(t *T) {
	c.t = t
}

// CopyFrom sets all patcher to the target.
func (c *Context[T]) CopyFrom(target *Context[T]) {
	c.SetLocked(target.GetLocked())
}
