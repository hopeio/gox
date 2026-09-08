package sync

import "sync"

type Pool[T any] struct {
	pool sync.Pool
}

// NewPool creates and returns a new instance.
func NewPool[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get gets a T from the pool, or creates a new one if the pool is empty.
//
// Note: do not Put a nil value of a non-pointer type T into the pool. The
// underlying sync.Pool returns such a value as a nil interface, and the type
// assertion below would panic for value types (e.g. Pool[int]).
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns x into the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}
