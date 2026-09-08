/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	reflectx "github.com/hopeio/gox/reflect"
	"golang.org/x/exp/constraints"

	"reflect"
	"unsafe"
)

// Every reports whether the condition holds.
func Every[S ~[]T, T any](slice S, fn func(T) bool) bool {
	for _, t := range slice {
		if !fn(t) {
			return false
		}
	}
	return true
}

// Some returns the value.
func Some[S ~[]T, T any](slice S, fn func(T) bool) bool {
	for _, t := range slice {
		if fn(t) {
			return true
		}
	}
	return false
}

// Zip returns a slice of pairs zipping s1 and s2 up to the shorter length.
// If the two slices have different lengths, the extra elements of the longer
// slice are ignored instead of causing an out-of-range panic.
func Zip[S ~[]T, T any](s1, s2 S) [][2]T {
	var newSlice [][2]T
	n := len(s1)
	if len(s2) < n {
		n = len(s2)
	}
	for i := 0; i < n; i++ {
		newSlice = append(newSlice, [2]T{s1[i], s2[i]})
	}
	return newSlice
}

// Deduplicate returns the slice with duplicate elements removed, preserving the
// order of first appearance.
func Deduplicate[S ~[]T, T comparable](slice S) S {
	newslice := make(S, 0, len(slice))
	set := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			newslice = append(newslice, v)
		}
	}
	return newslice
}

// ToMap converts the value.
func ToMap[S ~[]T, T any, K comparable, V any](s S, getKV func(T) (K, V)) map[K]V {
	m := make(map[K]V)
	for _, s := range s {
		k, v := getKV(s)
		m[k] = v
	}
	return m
}

// Classify returns the result.
func Classify[S ~[]T, T any, K comparable, V any](s S, getKV func(T) (K, V)) map[K][]V {
	m := make(map[K][]V)
	for _, s := range s {
		k, v := getKV(s)
		if ms, ok := m[k]; ok {
			m[k] = append(ms, v)
		} else {
			m[k] = []V{v}
		}

	}
	return m
}

// ForEach performs the operation.
func ForEach[S ~[]T, T any](s S, fn func(idx int, v T)) {
	for i, t := range s {
		fn(i, t)
	}
}

// ForEachValue performs the operation.
func ForEachValue[S ~[]T, T any](s S, fn func(v T)) {
	for _, v := range s {
		fn(v)
	}
}

// ForEachIndex performs the operation.
func ForEachIndex[S ~[]T, T any](s S, fn func(i int)) {
	for i := range s {
		fn(i)
	}
}

// ReverseForEach performs the operation.
func ReverseForEach[S ~[]T, T any](s S, fn func(idx int, v T)) {
	for i := len(s) - 1; i >= 0; i-- {
		fn(i, s[i])
	}
}

// Map returns the result.
func Map[T1S ~[]T1, T1, T2 any](s T1S, fn func(T1) T2) []T2 {
	ret := make([]T2, 0, len(s))
	for _, s := range s {
		ret = append(ret, fn(s))
	}
	return ret
}

// Filter returns the result.
func Filter[S ~[]T, T any](fn func(T) bool, src S) S {
	var dst S
	for _, v := range src {
		if fn(v) {
			dst = append(dst, v)
		}
	}
	return dst
}

// Reduce folds fn over the slice from left to right, using the first element as
// the initial accumulator. It returns the zero value for an empty slice instead
// of panicking on slices[1].
func Reduce[S ~[]T, T any](slices S, fn func(T, T) T) T {
	if len(slices) == 0 {
		var zero T
		return zero
	}
	ret := slices[0]
	for i := 1; i < len(slices); i++ {
		ret = fn(ret, slices[i])
	}
	return ret
}

// Convert converts the value.
//
// The conversion is memory-safe. Identical element types are reinterpreted with
// zero cost. When the two element types have the same size and contain no Go
// pointers, the backing array is reinterpreted in place (zero allocation), which
// is the intended fast path for, e.g., []convInt -> []int. In all other cases
// the elements are converted one by one. The previous code reinterpreted the
// backing array for any pair of equal Kind without checking the element size,
// which was undefined behavior / out-of-bounds for differing sizes (e.g.
// []int8 -> []int64).
func Convert[T1S ~[]T1, T2S ~[]T2, T1, T2 any](s T1S) T2S {
	t1, t2 := new(T1), new(T2)
	t1type, t2type := reflect.TypeOf(t1).Elem(), reflect.TypeOf(t2).Elem()
	if t1type == t2type {
		return any(s).(T2S)
	}
	if t1type.Size() == t2type.Size() && canReinterpret(t1type) && canReinterpret(t2type) {
		return unsafe.Slice((*T2)(unsafe.Pointer(unsafe.SliceData(s))), len(s))
	}
	if t1type.ConvertibleTo(t2type) {
		return Map(s, func(v T1) T2 {
			return reflect.ValueOf(v).Convert(t2type).Interface().(T2)
		})
	}
	if t2type.Kind() == reflect.Interface && t1type.Implements(t2type) {
		return Map(s, func(v T1) T2 {
			return reflect.ValueOf(v).Convert(t2type).Interface().(T2)
		})
	}
	if _, ok := any(t1).(T2); ok {
		return Map(s, func(v T1) T2 { return any(v).(T2) })
	}
	if _, ok := any(t2).(T1); ok {
		return Map(s, func(v T1) T2 { return any(v).(T2) })
	}
	panic("unsupported type")
}

// canReinterpret reports whether a value of type t can be safely reinterpreted as
// a different but same-sized type without confusing the garbage collector.
// Types that contain Go pointers must not be reinterpreted this way.
func canReinterpret(t reflect.Type) bool {
	switch t.Kind() {
	// String is {ptr,len}; reinterpreting it as a same-sized numeric type
	// (or vice versa) confuses the GC and is undefined behavior.
	case reflect.String, reflect.Pointer, reflect.UnsafePointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return false
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if !canReinterpret(t.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Array:
		return canReinterpret(t.Elem())
	default:
		return true
	}
}

// GuardSlice performs the operation.
func GuardSlice(buf *[]byte, n int) {
	c := cap(*buf)
	l := len(*buf)
	if c-l < n {
		c = c>>1 + n + l
		if c < 32 {
			c = 32
		}
		tmp := make([]byte, l, c)
		copy(tmp, *buf)
		*buf = tmp
	}
}

//go:nosplit
func PtrToSlicePtr(s unsafe.Pointer, l int, c int) unsafe.Pointer {
	slice := &reflectx.Slice{
		Ptr: s,
		Len: l,
		Cap: c,
	}
	return unsafe.Pointer(slice)
}

// FilterPlace returns the result.
func FilterPlace[S ~[]T, T any](slices S, fn func(T) bool) S {
	n := len(slices) - 1
	for i := 0; i <= n; {
		if fn(slices[i]) {
			if i < n {
				slices[i], slices[n] = slices[n], slices[i]
			}
			n--
			continue
		}
		i++
	}
	return slices[:n+1]
}

// Remove deletes the element at index i. An out-of-range index leaves the slice
// unchanged instead of panicking.
func Remove[S ~[]T, T any](slices S, i int) S {
	if i < 0 || i >= len(slices) {
		return slices
	}
	return append(slices[:i], slices[i+1:]...)
}

// TwoDimensionalSlice extracts the sub-region [rowStart,rowEnd) x
// [colStart,colEnd) of s. Out-of-range requests are clamped to valid bounds
// instead of panicking; a row whose column range is invalid yields a nil slice.
func TwoDimensionalSlice[S ~[][]T, T any](s S, rowStart, rowEnd, colStart, colEnd int) S {
	if rowStart < 0 || rowEnd > len(s) || rowStart > rowEnd {
		return S{}
	}
	ret := make([][]T, rowEnd-rowStart)
	for i := range ret {
		row := s[rowStart+i]
		if colStart < 0 || colEnd > len(row) || colStart > colEnd {
			ret[i] = nil
			continue
		}
		ret[i] = row[colStart:colEnd]
	}
	return ret
}

// ThreeDimensionalSlice extracts a sub-cube of s. Out-of-range requests are
// clamped to valid bounds instead of panicking; an invalid range on any axis
// yields a nil slice for that cell.
func ThreeDimensionalSlice[S ~[][][]T, T any](s S, rowStart, rowEnd, colStart, colEnd, sliceStart, sliceEnd int) S {
	if rowStart < 0 || rowEnd > len(s) || rowStart > rowEnd {
		return S{}
	}
	ret := make(S, rowEnd-rowStart)
	for i := range ret {
		plane := s[rowStart+i]
		if colStart < 0 || colEnd > len(plane) || colStart > colEnd {
			ret[i] = nil
			continue
		}
		ret[i] = make([][]T, colEnd-colStart)
		for j := range ret[i] {
			row := plane[colStart+j]
			if sliceStart < 0 || sliceEnd > len(row) || sliceStart > sliceEnd {
				ret[i][j] = nil
				continue
			}
			ret[i][j] = row[sliceStart:sliceEnd]
		}
	}
	return ret
}

// ToPtrs converts the value.
func ToPtrs[S ~[]T, T any](s S) []*T {
	ret := make([]*T, len(s))
	for i := range s {
		ret[i] = &s[i]
	}
	return ret
}

// Copy returns the result.
func Copy[S ~[]T, T any](s S) S {
	c := make([]T, len(s))
	copy(c, s)
	return c
}

// GroupBy returns the result.
func GroupBy[S ~[]T, T any, K comparable](s S, getK func(T) K) map[K][]T {
	m := make(map[K][]T)
	for _, s := range s {
		k := getK(s)
		if ms, ok := m[k]; ok {
			m[k] = append(ms, s)
		} else {
			m[k] = []T{s}
		}
	}
	return m
}

// Sum returns the result.
func Sum[S ~[]T, T constraints.Ordered](s S) T {
	var ret T
	for _, s := range s {
		ret += s
	}
	return ret
}
