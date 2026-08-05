package queue

import "sync"

type MutexQueue[T any] struct {
	v  []T
	mu sync.RWMutex
}

// NewMutexQueue creates and returns a new instance.
func NewMutexQueue[T any]() *MutexQueue[T] {
	return &MutexQueue[T]{v: make([]T, 0)}
}

// Enqueue updates or inserts a value.
func (q *MutexQueue[T]) Enqueue(v T) {
	q.mu.Lock()
	q.v = append(q.v, v)
	q.mu.Unlock()
}

// Dequeue removes or resets state.
func (q *MutexQueue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	if len(q.v) == 0 {
		q.mu.Unlock()
		return *new(T), false
	}
	v := q.v[0]
	q.v = q.v[1:]
	q.mu.Unlock()
	return v, true
}
