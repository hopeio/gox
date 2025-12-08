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

// Filter keep elements which satisfy the Predicate.
// 保留满足断言的元素
func Filter[T any](seq iter.Seq[T], test types.Predicate[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if test(v) && !yield(v) {
				return
			}
		}
	}
}

// Map transform the element use Fuction.
// 使用输入函数对每个元素进行转换
func Map[T, R any](seq iter.Seq[T], f types.UnaryFunction[T, R]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// FlatMap transform each element in Seq[T] to a new Seq[R].
// 将原本序列中的每个元素都转换为一个新的序列，
// 并将所有转换后的序列依次连接起来生成一个新的序列
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

// Peek visit every element in the Seq and leave them on the Seq.
// 访问序列中的每个元素而不消费它
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

// Distinct remove duplicate elements.
// 对序列中的元素去重
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

// Sorted sort elements in the Seq by Comparator.
// 对序列中的元素排序
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

// IsSorted
// 对序列中的元素是否排序
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

// Limit limits the number of elements in Seq.
// 限制元素个数
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

// Skip drop some elements of the Seq.
// 跳过指定个数的元素
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

// ForEach consume every elements in the Seq.
// 消费序列中的每个元素
func ForEach[T any](seq iter.Seq[T], accept types.Consumer[T]) {
	for v := range seq {
		accept(v)
	}
}

func Every[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if !test(v) {
			return false
		}
	}
	return true
}

func Some[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if test(v) {
			return true
		}
	}
	return false
}

// AllMatch test if every element are all match the Predicate.
// 是否每个元素都满足条件 == Every
func AllMatch[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if !test(v) {
			return false
		}
	}
	return true
}

// AnyMatch test if any element matches the Predicate.
// 是否有任意元素满足条件 == Some
func AnyMatch[T any](seq iter.Seq[T], test types.Predicate[T]) bool {
	for v := range seq {
		if test(v) {
			return true
		}
	}
	return false
}

// Reduce accumulate each element using the binary operation.
// 使用给定的累加函数, 累加序列中的每个元素.
// 序列中可能没有元素因此返回的是 Optional
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

// Fold accumulate each element using the BinaryFunction
// starting from the initial value.
// 从初始值开始, 通过 acc 函数累加每个元素
func Fold[T, R any](seq iter.Seq[T], initVal R, acc types.BinaryFunction[R, T, R]) (result R) {
	result = initVal
	for v := range seq {
		result = acc(result, v)
	}
	return result
}

// First find the first element in the Seq.
// 返回序列中的第一个元素(如有).
func First[T any](seq iter.Seq[T]) (T, bool) {
	for v := range seq {
		return v, true
	}
	return *new(T), false
}

// Count return the count of elements in the Seq.
// 返回序列中的元素个数
func Count[T any](seq iter.Seq[T]) (count int) {
	for _ = range seq {
		count++
	}
	return
}

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

func Operator[T any](seq iter.Seq[T], add types.BinaryOperator[T]) T {
	var result T
	var idx int
	for v := range seq {
		if idx == 0 {
			result = v
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
	var ok bool
	var i int
	for v := range it {
		if i == index {
			return v, true
		}
	}
	return zero, ok
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

func JoinBy[T any](it iter.Seq[T], toString func(T) string, sep string) string {
	var b strings.Builder
	for v := range it {
		b.WriteString(toString(v))
		b.WriteString(sep)
	}
	return b.String()[:b.Len()-len(sep)]
}
