/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForEachByIdx(t *testing.T) {
	val1 := []string{"a", "b", "c"}
	ForEachIndex(val1, func(i int) {
		fmt.Println(val1[i])
	})
}

type Int int
type Int8 int8

func (Int8) Interface() {
}

func (Int8) Interface2() {
}

type Struct1 struct {
	Field1 int
}

type Struct2 struct {
	Field2 int
}

type Interface interface {
	Interface()
}

type Interface2 interface {
	Interface2()
}

func TestCast(t *testing.T) {
	var s1 Struct1
	var s2 Struct2
	fmt.Println("AssignableTo:", reflect.TypeOf(s1).AssignableTo(reflect.TypeOf(s2)))
	fmt.Println("ConvertibleTo:", reflect.TypeOf(s1).ConvertibleTo(reflect.TypeOf(s2)))
	var x *Int
	var y *int

	// 获取 *Int 和 *int 的底层类型
	fmt.Println("AssignableTo:", reflect.TypeOf(x).AssignableTo(reflect.TypeOf(y)))
	fmt.Println("ConvertibleTo:", reflect.TypeOf(x).ConvertibleTo(reflect.TypeOf(y)))
	val1 := []Int{1, 2, 3}
	val2 := Convert[[]Int, []int](val1)
	t.Log(val2)
	val3 := []Int8{1, 2, 3}
	val4 := Convert[[]Int8, []int](val3)
	t.Log(val4)

	val5 := []Int8{1, 2, 3}
	val6 := Convert[[]Int8, []Interface](val5)
	t.Log(val6)

	val7 := []Interface{Int8(1), Int8(2), Int8(3)}
	val8 := Convert[[]Interface, []Int8](val7)
	t.Log(val8)

	val9 := Convert[[]Interface, []Interface2](val7)
	t.Log(val9)

	//v1 := Int8(2)
	//v2 := Interface(Int8(2))
	//v3 := Interface(v1)
	//v4 := v2.(Int8)
	//v5 := v2.(Interface)
	//v6 := Interface(v1)
	//v7:= any(Int8(1)).(Interface)
	//v8:= Interface(any(Int8(1)))
}

func TestFilterPlace(t *testing.T) {
	assert.Equal(t, []int{1, 3}, FilterPlace([]int{1, 2, 3}, func(i int) bool { return i == 2 }))
}
