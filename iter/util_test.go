/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"github.com/stretchr/testify/assert"
	"iter"
	"slices"
	"testing"
)

func TestIter(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	seq := slices.Values(s)
	t.Log(First[int](seq))
	for v := range seq {
		t.Log(v)
	}
	assert.Equal(t, true, StreamOf(slices.Values(s)).IsSorted(func(i int, i2 int) int {
		return i - i2
	}))
	var count int
	StreamOf(slices.Values(s)).Filter(func(i int) bool {
		return i%2 == 0
	}).Map(func(i int) int {
		return i + 10
	}).Peek(func(i int) {
		count++
	}).ForEach(func(i int) {
		t.Log(i)
	})
	t.Log(count)
}

func TestDistinct(t *testing.T) {
	s := []int{1, 2, 2, 5, 5, 6, 5}
	seq := Distinct(iter.Seq[int](slices.Values(s)), func(i int) int {
		return i
	})
	var times int
	for v := range seq {
		if v == 5 {
			if times == 1 {
				break
			}
			times++
		}
		t.Log(v)
	}
	StreamOf(slices.Values(s)).Distinct(func(i int) int {
		return i
	}).ForEach(func(i int) {
		t.Log(i)
	})
}
