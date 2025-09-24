/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
	"time"
)

type heap []int

func (h heap) Less(i, j int) bool {
	return h[i] < h[j]
}

func TestNewHeap(t *testing.T) {
	var arr heap
	for i := 0; i < 10; i++ {
		arr = append(arr, rand.Intn(10000))
	}
	heap := NewFromArray(arr)
	fmt.Println(heap)
	for i := 0; i < 10; i++ {
		heap.Push(rand.Intn(10000))
	}
	fmt.Println(heap)
	n := len(heap)
	for i := 0; i < n; i++ {
		t.Log(heap.Pop())
	}
}

type Foo struct {
	A int
	B int
}

type Foos []Foo

func (f Foos) Less(i, j int) bool {
	if f[i].A == f[j].A {
		return f[i].B < f[j].B
	}
	return f[i].A < f[j].A
}

func TestPushHeap(t *testing.T) {
	var arr Foos
	heap := NewFromArray(arr)
	heap.Push(Foo{10, 10})
	heap.Push(Foo{5, 5})
	heap.Push(Foo{8, 8})
	heap.Push(Foo{2, 2})
	heap.Push(Foo{5, 51})
	heap.Push(Foo{26, 26})
	heap.Push(Foo{6, 6})
	heap.Push(Foo{9, 9})
	heap.Push(Foo{5, 52})
	heap.Push(Foo{1, 1})
	for _, foo := range heap {
		t.Log(foo)
	}
	t.Log("------------------")
	n := len(heap)
	for i := 0; i < n; i++ {
		t.Log(heap.Pop())
	}
}

type floats []float64

func (fs floats) Less(i, j int) bool {
	return fs[i] < fs[j]
}

var data, sorted = func() (floats, []float64) {
	rand.Seed(time.Now().UnixNano())
	var data []float64
	for i := 0; i < 100; i++ {
		data = append(data, rand.Float64()*100)
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return data, sorted
}()

func TestMaintainsPriorityTinyQueue(t *testing.T) {
	q := New[float64, floats](0)
	for i := 0; i < len(data); i++ {
		q.Push(data[i])
	}
	v, _ := q.First()
	assert.Equal(t, v, sorted[0])
	var result []float64
	for len(q) > 0 {
		v, _ = q.Pop()
		result = append(result, v)
	}
	assert.Equal(t, result, sorted)
}

func TestAcceptsDataInConstructor(t *testing.T) {
	q := NewFromArray(data)
	var result []float64
	for len(q) > 0 {
		v, _ := q.Pop()
		result = append(result, v)
	}
	assert.Equal(t, result, sorted)
}
func TestHandlesEdgeCasesWithFewElements(t *testing.T) {
	q := New[float64, floats](0)
	q.Push(2)
	q.Push(1)
	q.Pop()
	q.Pop()
	q.Pop()
	q.Push(2)
	q.Push(1)
	v, _ := q.Pop()
	assert.Equal(t, 1.0, v)
	v, _ = q.Pop()
	assert.Equal(t, 2.0, v)
	_, ok := q.Pop()
	assert.Equal(t, false, ok)
}
