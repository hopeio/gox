/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package set

import (
	"iter"

	"github.com/hopeio/gox/maps"
	"github.com/hopeio/gox/types"
)

type Set[K comparable] map[K]struct{}

// New creates a new instance.
func New[K comparable]() Set[K] {
	return make(Set[K])
}

// Add updates or inserts a value.
func (s Set[K]) Add(key K) {
	s[key] = struct{}{}
}

// Contains reports whether the condition holds.
func (s Set[K]) Contains(key K) bool {
	_, ok := s[key]
	return ok
}

// Remove removes or resets state.
func (s Set[K]) Remove(key K) {
	delete(s, key)
}

// ToSlice converts the value.
func (s Set[K]) ToSlice() []K {
	return maps.Keys(s)
}

// Len returns the number of elements.
func (s Set[K]) Len() int {
	return len(s)
}

// Clear removes all elements.
func (s Set[K]) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// ForEach applies f to every element.
func (s Set[K]) ForEach(f types.Consumer[K]) {
	for k := range s {
		f(k)
	}
}

// All reports whether every element satisfies pred.
func (s Set[K]) All(pred types.Predicate[K]) bool {
	for k := range s {
		if !pred(k) {
			return false
		}
	}
	return true
}

// Any reports whether at least one element satisfies pred.
func (s Set[K]) Any(pred types.Predicate[K]) bool {
	for k := range s {
		if pred(k) {
			return true
		}
	}
	return false
}

// Filter returns a new set with elements that satisfy pred.
func (s Set[K]) Filter(pred types.Predicate[K]) Set[K] {
	r := New[K]()
	for k := range s {
		if pred(k) {
			r.Add(k)
		}
	}
	return r
}

// Map maps each element to a new comparable key, yielding a new set.
// Method-level generic (Go 1.27): the result type R is independent of K.
func (s Set[K]) Map[R comparable](f types.UnaryFunction[K, R]) Set[R] {
	r := New[R]()
	for k := range s {
		r.Add(f(k))
	}
	return r
}

// MapToSlice maps each element to a value of arbitrary type R.
func (s Set[K]) MapToSlice[R any](f types.UnaryFunction[K, R]) []R {
	r := make([]R, 0, len(s))
	for k := range s {
		r = append(r, f(k))
	}
	return r
}

// MapToMap maps each element to a (key, value) pair, yielding a new map.
func (s Set[K]) MapToMap[R comparable, V any](f types.UnaryFunction[K, types.Pair[R, V]]) map[R]V {
	r := make(map[R]V, len(s))
	for k := range s {
		rk, rv := f(k).First, f(k).Second
		r[rk] = rv
	}
	return r
}

// Union returns a new set with elements present in either set.
func (s Set[K]) Union(other Set[K]) Set[K] {
	r := New[K]()
	for k := range s {
		r.Add(k)
	}
	for k := range other {
		r.Add(k)
	}
	return r
}

// Intersect returns a new set with elements present in both sets.
func (s Set[K]) Intersect(other Set[K]) Set[K] {
	r := New[K]()
	for k := range s {
		if other.Contains(k) {
			r.Add(k)
		}
	}
	return r
}

// Difference returns a new set with elements in s but not in other.
func (s Set[K]) Difference(other Set[K]) Set[K] {
	r := New[K]()
	for k := range s {
		if !other.Contains(k) {
			r.Add(k)
		}
	}
	return r
}

// Seq yields all elements (satisfies iter.Seq[K]).
func (s Set[K]) Seq() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range s {
			if !yield(k) {
				return
			}
		}
	}
}
