/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package smap

import "sync"

type MutexMap[K comparable, V any] struct {
	m map[K]V
	l sync.RWMutex
}

// Set ...
func (s *MutexMap[K, V]) Set(key K, value V) {
	s.l.Lock()
	defer s.l.Unlock()
	s.m[key] = value
}

// Get ...
func (s *MutexMap[K, V]) Get(key K) (V, bool) {
	s.l.RLock()
	defer s.l.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

// New ...
func New[K comparable, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{m: make(map[K]V)}
}
