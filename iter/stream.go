/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"iter"
	"slices"

	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/types"
)

// NewStream returns the result.
func NewStream[T any](seq iter.Seq[T]) Stream[T] {
	return Stream[T](seq)
}

type Stream[T any] iter.Seq[T]

// Seq returns the result.
func (it Stream[T]) Seq() iter.Seq[T] {
	return iter.Seq[T](it)
}

// Filter returns the result.
func (it Stream[T]) Filter(test types.Predicate[T]) Stream[T] {
	return Stream[T](Filter(iter.Seq[T](it), test))
}

// Map returns the result.
func (it Stream[T]) Map[R any](f types.UnaryFunction[T, R]) Stream[R] {
	return Stream[R](Map(iter.Seq[T](it), f))
}

// FlatMap returns the result.
func (it Stream[T]) FlatMap[R any](f types.UnaryFunction[T, iter.Seq[R]]) Stream[R] {
	return Stream[R](FlatMap(iter.Seq[T](it), f))
}

// Peek returns the result.
func (it Stream[T]) Peek(accept types.Consumer[T]) Stream[T] {
	return Stream[T](Peek(iter.Seq[T](it), accept))
}

// Distinct returns the result.
func (it Stream[T]) Distinct(f types.UnaryFunction[T, int]) Stream[T] {
	return Stream[T](Distinct(iter.Seq[T](it), f))
}

// Sorted returns the result.
func (it Stream[T]) Sorted(cmp types.Comparator[T]) Stream[T] {
	return Stream[T](Sorted(iter.Seq[T](it), cmp))
}

// IsSorted reports whether the condition holds.
func (it Stream[T]) IsSorted(cmp types.Comparator[T]) bool {
	return IsSorted(iter.Seq[T](it), cmp)
}

// Limit returns the result.
func (it Stream[T]) Limit(limit int) Stream[T] {
	return Stream[T](Limit(iter.Seq[T](it), limit))
}

// Until returns the result.
func (it Stream[T]) Until(test types.Predicate[T]) Stream[T] {
	return Stream[T](Until(iter.Seq[T](it), test))
}

// Skip returns the result.
func (it Stream[T]) Skip(skip int) Stream[T] {
	return Stream[T](Skip(iter.Seq[T](it), skip))
}

// ForEach performs the operation.
func (it Stream[T]) ForEach(accept types.Consumer[T]) {
	ForEach(iter.Seq[T](it), accept)
}

// Collect returns the result.
func (it Stream[T]) Collect() []T {
	return slices.Collect(iter.Seq[T](it))
}

// All reports whether the condition holds.
func (it Stream[T]) All(test types.Predicate[T]) bool {
	return AllMatch(iter.Seq[T](it), test)
}

// Any reports whether the condition holds.
func (it Stream[T]) Any(test types.Predicate[T]) bool {
	return AnyMatch(iter.Seq[T](it), test)
}

// Reduce performs the operation.
func (it Stream[T]) Reduce(acc types.BinaryOperator[T]) (T, bool) {
	return Reduce(iter.Seq[T](it), acc)
}

// Fold returns the result.
func (it Stream[T]) Fold(initVal T, acc types.BinaryOperator[T]) T {
	return Fold(iter.Seq[T](it), initVal, types.BinaryFunction[T, T, T](acc))
}

// First performs the operation.
func (it Stream[T]) First() (T, bool) {
	return First(iter.Seq[T](it))
}

// Count returns the number of elements.
func (it Stream[T]) Count() int {
	return Count(iter.Seq[T](it))
}

// Sum returns the result.
func (it Stream[T]) Sum(add types.BinaryOperator[T]) T {
	return Operator(iter.Seq[T](it), add)
}

// Iter returns the result.
func (it Stream[T]) Iter() Iterator[T] {
	next, stop := iter.Pull(iter.Seq[T](it))
	return &seqIter[T]{next, stop}
}

// Zip pairs elements from this stream with another by index. Stops at the shorter one.
// Returns an iter.Seq (not Stream) to avoid recursive Stream[Pair[T,R]] instantiation.
func (it Stream[T]) Zip[R any](other iter.Seq[R]) iter.Seq[types.Pair[T, R]] {
	return Zip(iter.Seq[T](it), other)
}

// GroupBy groups elements by the key extracted from each element.
func (it Stream[T]) GroupBy[K comparable](key types.UnaryFunction[T, K]) map[K][]T {
	return GroupBy(iter.Seq[T](it), key)
}

// Partition splits elements into (matched, unmatched) by the predicate.
func (it Stream[T]) Partition(test types.Predicate[T]) ([]T, []T) {
	return Partition(iter.Seq[T](it), test)
}

// Chunk splits the stream into consecutive sub-slices of at most size elements.
// Returns an iter.Seq (not Stream) to avoid recursive Stream[[]T] instantiation.
func (it Stream[T]) Chunk(size int) iter.Seq[[]T] {
	return Chunk(iter.Seq[T](it), size)
}

// Window emits overlapping sliding windows of size elements.
// Returns an iter.Seq (not Stream) to avoid recursive Stream[[]T] instantiation.
func (it Stream[T]) Window(size int) iter.Seq[[]T] {
	return Window(iter.Seq[T](it), size)
}

// TakeWhile yields elements while the predicate holds, then stops.
func (it Stream[T]) TakeWhile(test types.Predicate[T]) Stream[T] {
	return Stream[T](TakeWhile(iter.Seq[T](it), test))
}

// DropWhile skips elements while the predicate holds, then yields the rest.
func (it Stream[T]) DropWhile(test types.Predicate[T]) Stream[T] {
	return Stream[T](DropWhile(iter.Seq[T](it), test))
}

// NoneMatch reports whether no element satisfies the predicate.
func (it Stream[T]) NoneMatch(test types.Predicate[T]) bool {
	return NoneMatch(iter.Seq[T](it), test)
}

// ForEachIndexed consumes each element with its index.
func (it Stream[T]) ForEachIndexed(accept func(int, T)) {
	ForEachIndexed(iter.Seq[T](it), accept)
}

// Find returns the first element satisfying the predicate.
func (it Stream[T]) Find(test types.Predicate[T]) (T, bool) {
	return Find(iter.Seq[T](it), test)
}

// FindLast returns the last element satisfying the predicate.
func (it Stream[T]) FindLast(test types.Predicate[T]) (T, bool) {
	return FindLast(iter.Seq[T](it), test)
}

// MinBy returns the minimum value by the less function.
func (it Stream[T]) MinBy(less cmp.LessFunc[T]) (T, bool) {
	return MinBy(iter.Seq[T](it), less)
}

// MaxBy returns the maximum value by the less function.
func (it Stream[T]) MaxBy(greater cmp.LessFunc[T]) (T, bool) {
	return MaxBy(iter.Seq[T](it), greater)
}
