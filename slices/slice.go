/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	"github.com/hopeio/utils/cmp"
	reflecti "github.com/hopeio/utils/reflect"
	"golang.org/x/exp/constraints"
	"slices"

	"reflect"
	"unsafe"
)

func Every[S ~[]T, T any](slice S, fn func(T) bool) bool {
	for _, t := range slice {
		if !fn(t) {
			return false
		}
	}
	return true
}

func Some[S ~[]T, T any](slice S, fn func(T) bool) bool {
	for _, t := range slice {
		if fn(t) {
			return true
		}
	}
	return false
}

func Zip[S ~[]T, T any](s1, s2 S) [][2]T {
	var newSlice [][2]T
	for i := range s1 {
		newSlice = append(newSlice, [2]T{s1[i], s2[i]})
	}
	return newSlice
}

// 去重
func Deduplicate[S ~[]T, T comparable](slice S) S {
	if len(slice) < SmallArrayLen {
		newslice := make(S, 0, 2)
		for i := 0; i < len(slice); i++ {
			if !slices.Contains(newslice, slice[i]) {
				newslice = append(newslice, slice[i])
			}
		}
		return newslice
	}
	set := make(map[T]struct{})
	for i := 0; i < len(slice); i++ {
		set[slice[i]] = struct{}{}
	}
	newslice := make(S, 0, len(slice))
	for k := range set {
		newslice = append(newslice, k)
	}
	return newslice
}

// Deprecated: use std slices.Contains
func Contains[S ~[]T, T comparable](s S, v T) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == v {
			return true
		}
	}
	return false
}

func ContainsByKey[S ~[]E, E cmp.EqualKey[K], K comparable](s S, v K) bool {
	for i := range s {
		if s[i].EqualKey() == v {
			return true
		}
	}
	return false
}

func Reverse[S ~[]T, T any](s S) S {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

func Max[S ~[]T, T constraints.Ordered](s S) T {
	if len(s) == 0 {
		return *new(T)
	}
	max := s[0]
	if len(s) == 1 {
		return max
	}
	for i := 1; i < len(s); i++ {
		if s[i] > max {
			max = s[i]
		}
	}

	return max
}

func Min[S ~[]T, T constraints.Ordered](s S) T {
	if len(s) == 0 {
		return *new(T)
	}
	min := s[0]
	if len(s) == 1 {
		return min
	}
	for i := 1; i < len(s); i++ {
		if s[i] < min {
			min = s[i]
		}
	}

	return min
}

// 将切片转换为map
func ToMap[S ~[]T, T any, K comparable, V any](s S, getKV func(T) (K, V)) map[K]V {
	m := make(map[K]V)
	for _, s := range s {
		k, v := getKV(s)
		m[k] = v
	}
	return m
}

// 将切片按照某个key分类
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

func Swap[S ~[]T, T any](s S, i, j int) {
	s[i], s[j] = s[j], s[i]
}

func ForEach[S ~[]T, T any](s S, handle func(idx int, v T)) {
	for i, t := range s {
		handle(i, t)
	}
}

func ForEachValue[S ~[]T, T any](s S, handle func(v T)) {
	for _, v := range s {
		handle(v)
	}
}

// 遍历切片,参数为下标，利用闭包实现遍历
func ForEachIndex[S ~[]T, T any](s S, handle func(i int)) {
	for i := range s {
		handle(i)
	}
}

func ReverseForEach[S ~[]T, T any](s S, handle func(idx int, v T)) {
	l := len(s)
	for i := l - 1; i > 0; i-- {
		handle(i, s[i])
	}
}

func Map[T1S ~[]T1, T1, T2 any](s T1S, fn func(T1) T2) []T2 {
	ret := make([]T2, 0, len(s))
	for _, s := range s {
		ret = append(ret, fn(s))
	}
	return ret
}

func Filter[S ~[]T, T any](fn func(T) bool, src S) S {
	var dst S
	for _, v := range src {
		if fn(v) {
			dst = append(dst, v)
		}
	}
	return dst
}

func Reduce[S ~[]T, T any](slices S, fn func(T, T) T) T {
	ret := fn(slices[0], slices[1])
	for i := 2; i < len(slices); i++ {
		ret = fn(ret, slices[i])
	}
	return ret
}

func Cast[T1S ~[]T1, T2S ~[]T2, T1, T2 any](s T1S) T2S {
	t1, t2 := new(T1), new(T2)
	t1type, t2type := reflect.TypeOf(t1).Elem(), reflect.TypeOf(t2).Elem()
	t1kind, t2kind := t1type.Kind(), t2type.Kind()

	if t1type.ConvertibleTo(t2type) && reflecti.CanCast(t1type, t2type, false) {
		if t1kind == t2kind {
			return unsafe.Slice((*T2)(unsafe.Pointer(unsafe.SliceData(s))), len(s))
		}
		if t1kind != reflect.Interface && t2kind != reflect.Interface {
			return Map(s, func(v T1) T2 { return *(*T2)(unsafe.Pointer(&v)) })
		}
	}

	if _, ok := any(t1).(T2); ok {
		return Map(s, func(v T1) T2 { return any(v).(T2) })
	}
	if _, ok := any(t2).(T1); ok {
		return Map(s, func(v T1) T2 { return any(v).(T2) })
	}
	panic("unsupported type")
}

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
	slice := &reflecti.Slice{
		Ptr: s,
		Len: l,
		Cap: c,
	}
	return unsafe.Pointer(slice)
}

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

func Remove[S ~[]T, T any](slices S, i int) S {
	return append(slices[:i], slices[i+1:]...)
}

func TwoDimensionalSlice[S ~[][]T, T any](s S, rowStart, rowEnd, colStart, colEnd int) S {
	ret := make([][]T, rowEnd-rowStart)
	for i := range ret {
		ret[i] = s[rowStart+i][colStart:colEnd]
	}
	return ret
}

func ThreeDimensionalSlice[S ~[][][]T, T any](s S, rowStart, rowEnd, colStart, colEnd, sliceStart, sliceEnd int) S {
	ret := make(S, rowEnd-rowStart)
	for i := range ret {
		ret[i] = make([][]T, colStart-colEnd)
		for j := range ret[i] {
			ret[i][j] = s[rowStart+i][colStart+j][sliceStart:sliceEnd]
		}
	}
	return ret
}

func ToPtrs[S ~[]T, T any](s S) []*T {
	ret := make([]*T, len(s))
	for i := range s {
		ret[i] = &s[i]
	}
	return ret
}
