/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package aop

import (
	"reflect"
	"unsafe"

	"bou.ke/monkey"
)

// Deprecated only support func var
func Invoke(before func(), target any, after func()) {
	v2 := reflect.ValueOf(target).Elem()
	if v2.Kind() != reflect.Func {
		panic("错误的类型")
	}

	oldFuncVal := reflect.MakeFunc(v2.Type(), nil)
	funcValuePtr := reflect.ValueOf(oldFuncVal).FieldByName("ptr").Pointer()
	funcPtr := (*Func)(unsafe.Pointer(funcValuePtr))
	funcPtr.codePtr = v2.Pointer()
	newFuncVal := reflect.MakeFunc(v2.Type(), func(in []reflect.Value) []reflect.Value {
		if before != nil {
			before()
		}
		if after != nil {
			defer after()
		}

		return oldFuncVal.Call(in)
	})
	v2.Set(newFuncVal)

}

type Func struct {
	codePtr uintptr
}

// TODO: 不可用,太深了,放弃
func aop(before func(), target any, after func()) {
	v2 := reflect.ValueOf(target)
	if v2.Kind() != reflect.Func {
		panic("错误的类型")
	}

	oldFuncVal := reflect.MakeFunc(v2.Type(), nil)
	funcValuePtr := reflect.ValueOf(oldFuncVal).FieldByName("ptr").Pointer()
	funcPtr := (*Func)(unsafe.Pointer(funcValuePtr))
	funcPtr.codePtr = v2.Pointer()
	newFuncVal := reflect.MakeFunc(v2.Type(), func(in []reflect.Value) []reflect.Value {
		if before != nil {
			before()
		}
		if after != nil {
			defer after()
		}

		return oldFuncVal.Call(in)
	})
	monkey.Patch(target, newFuncVal.Interface())
}

type value struct {
	_   uintptr
	ptr unsafe.Pointer
}
