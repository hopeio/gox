/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import "iter"

type Iterator[T any] interface {
	Next() (v T, ok bool)
}

type Iterable[T any] interface {
	Iter() Iterator[T]
}

type GoIter[T any] interface {
	Iterator[T]
	Stop()
}

// SeqIter returns the result.
func SeqIter[T any](seq iter.Seq[T]) Iterator[T] {
	next, stop := iter.Pull(seq)
	return seqIter[T]{next, stop}
}

type seqIter[T any] struct {
	next func() (T, bool)
	stop func()
}

// Next performs the operation.
func (a seqIter[T]) Next() (T, bool) {
	return a.next()
}

// Stop closes and releases resources.
func (a seqIter[T]) Stop() {
	a.stop()
}

// IterSeq returns the result.
func IterSeq[T any](iter Iterator[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			v, ok := iter.Next()
			if !ok || !yield(v) {
				return
			}
		}
	}
}

// Empty returns an iterator that yields no elements.
func Empty[T any]() iter.Seq[T] {
	return func(yield func(T) bool) {}
}

// Once yields a single element.
func Once[T any](v T) iter.Seq[T] {
	return func(yield func(T) bool) {
		yield(v)
	}
}

// Repeat yields the same value infinitely. Only safe to consume with a bounded
// operation (Limit/TakeWhile/etc).
func Repeat[T any](v T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(v) {
				return
			}
		}
	}
}

// RepeatN yields the same value n times (n <= 0 yields nothing).
func RepeatN[T any](v T, n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < n; i++ {
			if !yield(v) {
				return
			}
		}
	}
}

// Cycle repeats the source sequence infinitely. The source is materialized once
// into a slice. Only safe to consume with a bounded operation.
func Cycle[T any](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var data []T
		for v := range seq {
			data = append(data, v)
		}
		if len(data) == 0 {
			return
		}
		for {
			for _, v := range data {
				if !yield(v) {
					return
				}
			}
		}
	}
}
