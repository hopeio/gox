/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	"slices"

	"github.com/hopeio/gox"
	"github.com/hopeio/gox/cmp"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
)

// 没有泛型，范例，实际需根据不同类型各写一遍,用CmpKey，基本类型又用不了，go需要能给基本类型实现方法不能给外部类型实现方法
// 1.20以后字段均是comparable的结构体也是comparable的
// 判断是否有重合元素
func HasCoincide[S ~[]T, T comparable](s1, s2 S) bool {
	if len(s1) == 0 || len(s2) == 0 {
		return false
	}
	// 小数组
	if len(s1) < SmallArrayLen && len(s2) < SmallArrayLen {
		for i := range s1 {
			for j := range s2 {
				if s1[i] == s2[j] {
					return true
				}
			}
		}
	}
	// 同时遍历检测
	n, m := len(s1), len(s2)
	tmpMap := make(map[T]struct{})
	l := gox.TernaryOperator(n > m, n, m)
	for i := range l {
		if i < n {
			tmpMap[s1[i]] = struct{}{}
		}
		if i < m {
			if _, ok := tmpMap[s2[i]]; ok {
				return true
			}
		}
	}

	return false
}

func HasCoincideByKey[S ~[]E, E cmp.EqualKey[T], T comparable](s1, s2 S) bool {
	if len(s1) == 0 || len(s2) == 0 {
		return false
	}
	// 小数组
	if len(s1) < SmallArrayLen && len(s2) < SmallArrayLen {
		for i := range s1 {
			for j := range s2 {
				if s1[i].EqualKey() == s2[j].EqualKey() {
					return true
				}
			}
		}
	}

	// 同时遍历检测
	n, m := len(s1), len(s2)
	tmpMap := make(map[T]struct{})
	l := gox.TernaryOperator(n > m, n, m)
	for i := range l {
		if i < n {
			tmpMap[s1[i].EqualKey()] = struct{}{}
		}
		if i < m {
			if _, ok := tmpMap[s2[i].EqualKey()]; ok {
				return true
			}
		}
	}
	return false
}

func RemoveDuplicates[S ~[]T, T comparable](s S) S {
	if len(s) == 0 {
		return s
	}
	var m = make(map[T]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return maps.Keys(m)
}

// 默认保留先遍历到的
func RemoveDuplicatesByKey[S ~[]E, E cmp.EqualKey[T], T comparable](s S) S {
	if len(s) == 0 {
		return s
	}
	var m = make(map[T]E)
	for _, v := range s {
		key := v.EqualKey()
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = v
	}
	return maps.Values(m)
}

func RemoveDuplicatesByKeyRetainBehind[S ~[]E, E cmp.EqualKey[T], T comparable](s S) S {
	if len(s) == 0 {
		return s
	}
	var m = make(map[T]E)
	for _, i := range s {
		m[i.EqualKey()] = i
	}
	return maps.Values(m)
}

// 取交集
func Intersection[S ~[]T, T comparable](a S, b S) S {
	if len(a) == 0 || len(b) == 0 {
		return S{}
	}
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		if len(a) > len(b) {
			return smallArrayIntersection(a, b)
		}
		return smallArrayIntersection(b, a)
	}
	return intersection(a, b)
}

func smallArrayIntersection[S ~[]T, T comparable](a S, b S) S {
	var ret S
	for _, x := range a {
		if slices.Contains(b, x) {
			ret = append(ret, x)
		}
	}
	return ret
}

func intersection[S ~[]T, T comparable](a S, b S) S {
	return maps.Keys(IntersectionMap(a, b))
}

func IntersectionMap[S ~[]T, T comparable](a S, b S) map[T]struct{} {
	if len(a) == 0 || len(b) == 0 {
		return make(map[T]struct{})
	}
	aMap, intersectionMap := make(map[T]struct{}), make(map[T]struct{})
	short, long := a, b
	if len(a) > len(b) {
		short, long = b, a
	}
	for _, v := range short {
		aMap[v] = struct{}{}
	}
	// 这里可以用两个map，也可以用map[T]bool类型最后过滤出来,但是最后过滤相当于遍历了两遍a
	for _, v := range long {
		if _, ok := aMap[v]; ok {
			intersectionMap[v] = struct{}{}
		}
	}
	return intersectionMap
}

// 默认保留前一个的,靠前的元素
func IntersectionByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	if len(a) == 0 || len(b) == 0 {
		return S{}
	}
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		return smallArrayIntersectionByKey(a, b)
	}
	tmpMap, intersectionMap := make(map[T]struct{}), make(map[T]E)

	for _, v := range b {
		key := v.EqualKey()
		if _, ok := tmpMap[key]; ok {
			continue
		}
		tmpMap[key] = struct{}{}
	}
	// 这里可以用两个map，也可以用map[T]bool类型最后过滤出来,但是最后过滤相当于遍历了两遍a
	for _, v := range a {
		key := v.EqualKey()
		if _, ok := intersectionMap[key]; ok {
			continue
		}
		if _, ok := tmpMap[key]; ok {
			intersectionMap[key] = v
		}
	}
	return maps.Values(intersectionMap)
}

func smallArrayIntersectionByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	var ret S
	for _, x := range a {
		if slices.ContainsFunc(b, func(e E) bool {
			return x.EqualKey() == e.EqualKey()
		}) {
			ret = append(ret, x)
		}
	}
	return ret
}

// 有序数组取交集
func OrderedArrayIntersection[S ~[]T, T constraints.Ordered](a S, b S) S {
	if len(a) == 0 || len(b) == 0 {
		return S{}
	}
	var ret S
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	var idx int
	for _, x := range a {
		if x > b[len(b)-1] {
			return ret
		}
		for j := idx; idx < len(b)-1; j++ {
			if a[len(a)-1] < b[idx] {
				return ret
			}
			if x == b[idx] {
				ret = append(ret, x)
				idx = j
			}
		}
	}
	return ret
}

// 取并集
func Union[S ~[]T, T comparable](a S, b S) S {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	var set = make(map[T]struct{}, len(a)+len(b))
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, x := range b {
		set[x] = struct{}{}
	}
	return maps.Keys(set)
}

// 默认保留第一个的,靠前的
func UnionByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	var m = make(map[T]E, len(a)+len(b))
	for _, x := range a {
		key := x.EqualKey()
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = x
	}
	for _, x := range b {
		key := x.EqualKey()
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = x
	}
	return maps.Values(m)
}

// 取差集,返回为A-B
func DifferenceSet[S ~[]T, T comparable](a S, b S) S {
	if len(a) == 0 {
		return S{}
	}
	if len(b) == 0 {
		return a
	}
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		return smallArrayDifferenceSet(a, b)
	}
	return differenceSet(a, b)
}

func smallArrayDifferenceSet[S ~[]T, T comparable](a S, b S) S {
	var diff S
	for _, x := range a {
		if !slices.Contains(b, x) {
			diff = append(diff, x)
		}
	}
	return diff
}

func differenceSet[S ~[]T, T comparable](a S, b S) S {
	var diff S
	if len(b)/len(a) >= 2 {
		aMap := make(map[T]bool)
		for _, x := range a {
			aMap[x] = false
		}
		for _, x := range b {
			if _, ok := aMap[x]; ok {
				aMap[x] = true
			}
		}
		for v, exits := range aMap {
			if !exits {
				diff = append(diff, v)
			}
		}
	} else {
		bMap := make(map[T]struct{})
		for _, x := range b {
			bMap[x] = struct{}{}
		}
		for _, x := range a {
			if _, ok := bMap[x]; !ok {
				diff = append(diff, x)
			}
		}
	}
	return diff
}

// 指定key取差集,返回为A-B
func DifferenceSetByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	if len(a) == 0 {
		return S{}
	}
	if len(b) == 0 {
		return a
	}
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		return smallArrayDifferenceSetByKey(a, b)
	}
	return differenceSetByKey(a, b)
}

func smallArrayDifferenceSetByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	var diff S
	for _, x := range a {
		if !slices.ContainsFunc(b, func(e E) bool {
			return x.EqualKey() == e.EqualKey()
		}) {
			diff = append(diff, x)
		}
	}
	return diff
}

func differenceSetByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a S, b S) S {
	var diff S
	if len(b)/len(a) >= 2 {
		aMap := make(map[T]bool)
		for _, x := range a {
			aMap[x.EqualKey()] = false
		}
		for _, x := range b {
			if _, ok := aMap[x.EqualKey()]; ok {
				aMap[x.EqualKey()] = true
			}
		}
		for _, x := range a {
			if !aMap[x.EqualKey()] {
				diff = append(diff, x)
			}
		}

	} else {
		bMap := make(map[T]struct{})
		for _, x := range b {
			bMap[x.EqualKey()] = struct{}{}
		}
		for _, x := range a {
			if _, ok := bMap[x.EqualKey()]; !ok {
				diff = append(diff, x)
			}
		}
	}
	return diff
}

// 两个数组各自相对的差集,返回为A-B,B-A
func Difference[S ~[]T, T comparable](a, b S) (S, S) {
	if len(a) == 0 {
		return S{}, b
	}
	if len(b) == 0 {
		return a, S{}
	}
	var diff1, diff2 S
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		intersectionMap := make(map[T]struct{})
	Loop:
		for _, i := range a {
			for _, j := range b {
				if i == j {
					intersectionMap[i] = struct{}{}
					continue Loop
				}
			}
			diff1 = append(diff1, i)
		}
		for _, i := range b {
			if _, ok := intersectionMap[i]; !ok {
				diff2 = append(diff2, i)
			}
		}
		return diff1, diff2
	}
	aMap, bMap := make(map[T]struct{}), make(map[T]struct{})
	for _, v := range a {
		aMap[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := aMap[v]; !ok {
			diff2 = append(diff2, v)
		}
		bMap[v] = struct{}{}
	}
	for _, v := range a {
		if _, ok := bMap[v]; !ok {
			diff1 = append(diff1, v)
		}
	}
	return diff1, diff2
}

// 取差集，通过循环比较key
func DifferenceByKey[S ~[]E, E cmp.EqualKey[T], T comparable](a, b S) (S, S) {
	if len(a) == 0 {
		return S{}, b
	}
	if len(b) == 0 {
		return a, S{}
	}
	var diff1, diff2 S
	if len(a) < SmallArrayLen && len(b) < SmallArrayLen {
		intersectionMap := make(map[T]struct{})
	Loop:
		for _, i := range a {
			for _, j := range b {
				if i.EqualKey() == j.EqualKey() {
					intersectionMap[i.EqualKey()] = struct{}{}
					continue Loop
				}
			}
			diff1 = append(diff1, i)
		}
		for _, i := range b {
			if _, ok := intersectionMap[i.EqualKey()]; !ok {
				diff2 = append(diff2, i)
			}
		}
	}
	aMap, bMap := make(map[T]struct{}), make(map[T]struct{})
	for _, v := range a {
		aMap[v.EqualKey()] = struct{}{}
	}
	for _, v := range b {
		if _, ok := aMap[v.EqualKey()]; !ok {
			diff2 = append(diff2, v)
		}
		bMap[v.EqualKey()] = struct{}{}
	}
	for _, v := range a {
		if _, ok := bMap[v.EqualKey()]; !ok {
			diff1 = append(diff1, v)
		}
	}
	return diff1, diff2
}

// 交集和差集，返回A∪B A∩B,A-B,B-A
func UnionAndIntersectionAndDifference[S ~[]T, T comparable](a, b S) (S, S, S, S) {
	if len(a) == 0 {
		return b, b, S{}, b
	}
	if len(b) == 0 {
		return a, a, a, S{}
	}
	var diff1, diff2 S
	unionMap, intersectionMap, aMap, bMap := make(map[T]struct{}), make(map[T]struct{}), make(map[T]struct{}), make(map[T]struct{})

	for _, v := range a {
		aMap[v] = struct{}{}
		unionMap[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := aMap[v]; !ok {
			diff2 = append(diff2, v)
		} else {
			intersectionMap[v] = struct{}{}
		}
		bMap[v] = struct{}{}
		unionMap[v] = struct{}{}
	}
	for _, v := range a {
		if _, ok := bMap[v]; !ok {
			diff1 = append(diff1, v)
		}
	}

	return maps.Keys(unionMap), maps.Keys(intersectionMap), diff1, diff2
}
