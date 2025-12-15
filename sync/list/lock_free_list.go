// Copyright 2024 hopeio. All rights reserved.
// Licensed under the MIT License that can be found in the LICENSE file.
// @Created by jyb

package list

import (
	"sync/atomic"

	"github.com/hopeio/gox/sync"
)

type LockFreeList[T any] struct {
	head atomic.Pointer[sync.Node[T]]
	tail atomic.Pointer[sync.Node[T]]
	size uint64
}

func NewLockFreeList[T any]() *LockFreeList[T] {
	return &LockFreeList[T]{}
}

func (l *LockFreeList[T]) Push(v T) {
	node := &sync.Node[T]{Value: v}
	if atomic.LoadUint64(&l.size) == 0 {
		l.head.Store(node)
		l.tail.Store(node)
		atomic.AddUint64(&l.size, 1)
		return
	}
	var last, lastnext *sync.Node[T]
	for {
		last = l.tail.Load()
		lastnext = last.Next.Load()
		if l.tail.Load() == last { // are tail and next consistent?
			if lastnext == nil { // was tail pointing to the last node?
				if last.Next.CompareAndSwap(lastnext, node) { // try to link item at the end of linked list
					l.tail.CompareAndSwap(last, node) // enqueue is done. try swing tail to the inserted node
					atomic.AddUint64(&l.size, 1)
					return
				}
			} else { // tail was not pointing to the last node
				l.tail.CompareAndSwap(last, lastnext) // try swing tail to the next node
			}
		}
	}
}

func (l *LockFreeList[T]) Pop() (v T, ok bool) {
	var last, lastnext *sync.Node[T]
	for {
		last = l.tail.Load()
		lastnext = last.Next.Load()
		if l.tail.Load() == last { // are tail and next consistent?
			if lastnext == nil { // is tail pointing to the last node?
				return *new(T), false
			} else { // tail was not pointing to the last node
				l.tail.CompareAndSwap(last, lastnext) // try swing tail to the
			}
		}
	}
}

func (l *LockFreeList[T]) Len() uint64 {
	return atomic.LoadUint64(&l.size)
}
