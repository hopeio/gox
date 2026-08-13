/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"iter"
	"slices"
	"strings"

	"github.com/hopeio/gox/container"
	"github.com/hopeio/gox/types"
)

// Filter returns the result.
func Filter[T any](seq iter.Seq[T], test types.Predicate[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if test(v) && !yield(v) {
				return
			}
		}
	}
}

// Map returns the result.
func Map[T, R any](seq iter.Seq[T], f types.UnaryFunction[T, R]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// FlatMap returns the result.
func FlatMap[T, R any](seq iter.Seq[T], flatten types.UnaryFunction[T, iter.Seq[R]]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for v := range seq {
			for v2 := range flatten(v) {
				if !yield(v2) {
					return
				}
			}
		}
	}
}

// Peek returns the result.
func Peek[T any](seq iter.Seq[T], accept types.Consumer[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			accept(v)
			if !yield(v) {
				return
			}
		}
	}
}

// Distinct returns the result.
func Distinct[T any, C comparable](seq iter.Seq[T], f types.UnaryFunction[T, C]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var set = make(map[C]struct{})
		for v := range seq {
			k := f(v)
			_, ok := set[k]
			if !ok {
				if !yield(v) {
					return
				}
				set[k] = struct{}{}
			}
		}
	}
}

// Sorted returns the result.
func Sorted[T any](it iter.Seq[T], cmp types.Comparator[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		vals := slices.SortedFunc(it, cmp)
		for _, v := range vals {
			if !yield(v) {
				return
			}
		}
	}
}

// IsSorted reports whether the condition holds.
func IsSorted[T any](seq iter.Seq[T], cmp types.Comparator[T]) bool {
	var last T
	check := func(curr T) bool {
		if cmp(last, curr) >= 0 {
			return false
		}
		last = curr
		return true
	}

	var has bool
	for v := range seq {
		if !has {
			last = v
			has = true
		} else {
			if !check(v) {
				return false
			}
		}
	}
	return true
}

// Limit returns the result.
func Limit[T any](seq iter.Seq[T], limit int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			limit--
			if limit < 0 {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Skip returns the result.
func Skip[T any](seq iter.Seq[T], skip int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			skip--
			if skip < 0 {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// UntilComparable returns the result.
func UntilComparable[T comparable](seq iter.Seq[T], e T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if v == e {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Until returns the result.
func Until[T any](seq iter.Seq[T], match types.Predicate[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if match(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ForEach performs the operation.
func ForEach[T any](seq iter.Seq[T], accept types.Consumer[T]) {
	for v := range seq {
		accept(v)
	}
}

// Every reports whether the condition holds.
func Every[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if !test(v) {
			return false
		}
	}
	return true
}

// Some returns the value.
func Some[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if test(v) {
			return true
		}
	}
	return false
}

// AllMatch reports whether the condition holds.
func AllMatch[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if !test(v) {
			return false
		}
	}
	return true
}

// AnyMatch reports whether the condition holds.
func AnyMatch[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if test(v) {
			return true
		}
	}
	return false
}

// Reduce performs the operation.
func Reduce[T any](seq iter.Seq[T], acc types.BinaryOperator[T]) (T, bool) {
	var result T
	var has bool
	for v := range seq {
		if !has {
			result = v
			has = true
		} else {
			result = acc(result, v)
		}
	}
	if has {
		return result, has
	}
	return result, has
}

// Fold performs the operation.
func Fold[T, R any](seq iter.Seq[T], initVal R, acc types.BinaryFunction[R, T, R]) (result R) {
	result = initVal
	for v := range seq {
		result = acc(result, v)
	}
	return result
}

// First performs the operation.
func First[T any](seq iter.Seq[T]) (T, bool) {
	for v := range seq {
		return v, true
	}
	return *new(T), false
}

// Count returns the number of elements.
func Count[T any](seq iter.Seq[T]) (count int) {
	for _ = range seq {
		count++
	}
	return
}

// Enumerate returns the result.
func Enumerate[T any](seq iter.Seq[T]) iter.Seq[types.Pair[int, T]] {
	return func(yield func(types.Pair[int, T]) bool) {
		var count int
		for v := range seq {
			if !yield(types.PairOf(count, v)) {
				return
			}
			count++
		}
	}
}

// Chain returns the result.
func Chain[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Operator returns the result.
func Operator[T any](seq iter.Seq[T], add types.BinaryOperator[T]) T {
	var result T
	var idx int
	for v := range seq {
		// idx 必须无条件递增：曾在首元素分支 continue 跳过递增，归约从不执行
		if idx == 0 {
			result = v
			idx++
			continue
		}
		result = add(result, v)
		idx++
	}
	return result
}

// Ruturns true if the count of Iterator is 0.
func IsEmpty[T any](it iter.Seq[T]) bool {
	for _ = range it {
		return false
	}
	return true
}

// Ruturns true if the count of Iterator is 0.
func IsNotEmpty[T any](it iter.Seq[T]) bool {
	for _ = range it {
		return true
	}
	return false
}

// Returns true if the target is included in the iterator.
func Contains[T comparable](it iter.Seq[T], target T) bool {
	for v := range it {
		if v == target {
			return true
		}
	}
	return false
}

// OperatorBy returns the result.
func OperatorBy[T any](it iter.Seq[T], f types.BinaryOperator[T]) T {
	result, _ := Reduce(it, func(a, b T) T {
		return f(a, b)
	})
	return result
}

// Return the right element.
func Last[T any](it iter.Seq[T]) (T, bool) {
	var result T
	var ok bool
	for v := range it {
		if !ok {
			ok = true
		}
		result = v
	}
	return result, ok
}

// Return the element at index.
func At[T any](it iter.Seq[T], index int) (T, bool) {
	var zero T
	var i int
	for v := range it {
		if i == index {
			return v, true
		}
		i++
	}
	return zero, false
}

// Splitting an iterator whose elements are pair into two lists.
func Unzip[A any, B any](it iter.Seq[types.Pair[A, B]]) ([]A, []B) {
	var arrA = make([]A, 0)
	var arrB = make([]B, 0)
	for p := range it {
		arrA = append(arrA, p.First)
		arrB = append(arrB, p.Second)
	}
	return arrA, arrB
}

// to built-in map.
func ToMap[K comparable, V any](it iter.Seq[types.Pair[K, V]]) map[K]V {
	var r = make(map[K]V)
	for p := range it {
		r[p.First] = p.Second
	}
	return r
}

// ToSlice converts the value.
func ToSlice[V any](it iter.Seq[V]) []V {
	var r []V
	for p := range it {
		r = append(r, p)
	}
	return r
}

// Collect Collecting via Collector.
func Collect[T any, S any, R any](it iter.Seq[T], collector container.Collector[S, T, R]) R {
	var s = collector.Builder()
	for v := range it {
		collector.Append(s, v)
	}
	return collector.Finish(s)
}

// Merge returns the result.
func Merge[T any](iters ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, it := range iters {
			for v := range it {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// JoinBy returns the result.
func JoinBy[T any](it iter.Seq[T], toString func(T) string, sep string) string {
	var b strings.Builder
	for v := range it {
		b.WriteString(toString(v))
		b.WriteString(sep)
	}
	return b.String()[:b.Len()-len(sep)]
}
