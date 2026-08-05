/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ringqueue

type RingQueue[T any] struct {
	head, tail int
	len        int
	buf        []T
	zero       T
}

// New creates a new instance.
func New[T any](capacity int) *RingQueue[T] {
	nodes := make([]T, capacity)
	return &RingQueue[T]{
		head: -1,
		tail: -1,
		buf:  nodes,
	}
}

// Length returns the result.
func (q *RingQueue[T]) Length() int {
	return q.len
}

// Capacity returns the result.
func (q *RingQueue[T]) Capacity() int {
	return len(q.buf)
}

// Front performs the operation.
func (q *RingQueue[T]) Front() (T, bool) {
	if q.len == 0 {
		return q.zero, false
	}

	return q.buf[q.head], true
}

// Tail performs the operation.
func (q *RingQueue[T]) Tail() (T, bool) {
	if q.len == 0 {
		return q.zero, false
	}

	return q.buf[q.tail], true
}

// Enqueue updates or inserts a value.
func (q *RingQueue[T]) Enqueue(value T) bool {
	if q.IsFull() {
		return false
	}

	q.tail++
	if q.tail == len(q.buf) {
		q.tail = 0
	}
	q.buf[q.tail] = value
	q.len++

	if q.len == 1 {
		q.head = q.tail
	}

	return true
}

// Dequeue removes or resets state.
func (q *RingQueue[T]) Dequeue() (T, bool) {
	if q.len == 0 {
		return q.zero, false
	}

	result := q.buf[q.head]
	q.buf[q.head] = q.zero
	q.head++
	q.len--
	if q.head == len(q.buf) {
		q.head = 0
	}

	return result, true
}

// IsFull checks if the ring buffer is full
func (q *RingQueue[T]) IsFull() bool {
	return q.len == len(q.buf)
}

// LookAll reads all elements from ring buffer
// this method doesn't consume all elements
func (q *RingQueue[T]) LookAll() []T {
	all := make([]T, q.len)
	if q.len == 0 {
		return all
	}
	j := 0
	for i := q.head; ; i++ {
		if i == len(q.buf) {
			i = 0
		}
		all[j] = q.buf[i]
		if i == q.tail {
			break
		}
		j++
	}
	return all
}
