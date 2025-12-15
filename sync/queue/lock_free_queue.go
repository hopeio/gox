// Copyright 2019 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package queue

import (
	"sync/atomic"

	"github.com/hopeio/gox/sync"
)

// LockFreeQueue implements lock-free FIFO freelist based queue.
// ref: https://dl.acm.org/citation.cfm?doid=248052.248106
type LockFreeQueue[T any] struct {
	head atomic.Pointer[sync.Node[T]]
	tail atomic.Pointer[sync.Node[T]]
	len  uint64
	zero T
}

// NewLockFreeQueue creates a new lock-free queue.
func NewLockFreeQueue[T any]() *LockFreeQueue[T] {
	head := sync.Node[T]{} // allocate a free item
	queue := &LockFreeQueue[T]{}
	queue.head.Store(&head)
	queue.tail.Store(&head)
	return queue
}

// Enqueue puts the given value v at the tail of the queue.
func (q *LockFreeQueue[T]) Enqueue(v T) {
	i := &sync.Node[T]{Value: v} // allocate new item
	var last, lastnext *sync.Node[T]
	for {
		last = q.tail.Load()
		lastnext = last.Next.Load()
		if q.tail.Load() == last { // are tail and next consistent?
			if lastnext == nil { // was tail pointing to the last node?
				if last.Next.CompareAndSwap(lastnext, i) { // try to link item at the end of linked list
					q.tail.CompareAndSwap(last, i) // enqueue is done. try swing tail to the inserted node
					atomic.AddUint64(&q.len, 1)
					return
				}
			} else { // tail was not pointing to the last node
				q.tail.CompareAndSwap(last, lastnext) // try swing tail to the next node
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *LockFreeQueue[T]) Dequeue() (T, bool) {
	var first, last, firstnext *sync.Node[T]
	for {
		first = q.head.Load()
		last = q.tail.Load()
		firstnext = first.Next.Load()
		if first == q.head.Load() { // are head, tail and next consistent?
			if first == last { // is queue empty?
				if firstnext == nil { // queue is empty, couldn't dequeue
					return q.zero, false
				}
				q.tail.CompareAndSwap(last, firstnext) // tail is falling behind, try to advance it
			} else { // read value before cas, otherwise another dequeue might free the next node
				v := firstnext.Value
				if q.head.CompareAndSwap(first, firstnext) { // try to swing head to the next node
					atomic.AddUint64(&q.len, ^uint64(0))
					return v, true // queue was not empty and dequeue finished.
				}
			}
		}
	}
}

// Len returns the length of the queue.
func (q *LockFreeQueue[T]) Len() uint64 {
	return atomic.LoadUint64(&q.len)
}
