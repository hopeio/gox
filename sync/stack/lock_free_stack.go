// Copyright 2019 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package stack

import (
	"sync/atomic"

	"github.com/hopeio/gox/sync"
)

// LockFreeStack implements lock-free freelist based stack.
type LockFreeStack[T any] struct {
	top  atomic.Pointer[sync.Node[T]]
	size uint64
}

// NewLockFreeStack creates a new lock-free queue.
func NewLockFreeStack[T any]() *LockFreeStack[T] {
	return &LockFreeStack[T]{}
}

// Pop pops value from the top of the stack.
func (s *LockFreeStack[T]) Pop() (T, bool) {
	if atomic.LoadUint64(&s.size) == 0 {
		return *new(T), false
	}

	var top, next *sync.Node[T]
	var item *sync.Node[T]
	for {
		top = s.top.Load()
		if top == nil {
			return *new(T), false
		}
		item = top
		next = item.Next.Load()
		if s.top.CompareAndSwap(top, next) {
			atomic.AddUint64(&s.size, ^uint64(0))
			return item.Value, true
		}
	}
}

// Push pushes a value on top of the stack.
func (s *LockFreeStack[T]) Push(v T) {
	item := sync.Node[T]{Value: v}
	var top *sync.Node[T]
	for {
		top = s.top.Load()
		item.Next.Store(top)
		if s.top.CompareAndSwap(top, &item) {
			atomic.AddUint64(&s.size, 1)
			return
		}
	}
}

func (s *LockFreeStack[T]) Len() uint64 {
	return atomic.LoadUint64(&s.size)
}
