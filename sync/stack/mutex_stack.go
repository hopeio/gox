package stack

import "sync"

type MutexStack[T any] struct {
	v  []T
	mu sync.RWMutex
}

// NewMutexStack creates and returns a new instance.
func NewMutexStack[T any]() *MutexStack[T] {
	return &MutexStack[T]{v: make([]T, 0)}
}

// Push ...
func (s *MutexStack[T]) Push(v T) {
	s.mu.Lock()
	s.v = append(s.v, v)
	s.mu.Unlock()
}

// Pop ...
func (s *MutexStack[T]) Pop() (T, bool) {
	s.mu.Lock()
	if len(s.v) == 0 {
		s.mu.Unlock()
		return *new(T), false
	}
	v := s.v[len(s.v)-1]
	s.v = s.v[:len(s.v)-1]
	s.mu.Unlock()
	return v, true
}
