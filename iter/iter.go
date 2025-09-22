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

func SeqIter[T any](seq iter.Seq[T]) Iterator[T] {
	next, stop := iter.Pull(seq)
	return seqIter[T]{next, stop}
}

type seqIter[T any] struct {
	next func() (T, bool)
	stop func()
}

func (a seqIter[T]) Next() (T, bool) {
	return a.next()
}

func (a seqIter[T]) Stop() {
	a.stop()
}

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
