/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binary

import (
	"reflect"
	"unsafe"

	reflectx "github.com/hopeio/gox/reflect"
)

/*
 * This serialization operates on pointers. If a struct contains pointers, it
 * cannot be restored from []byte and is only useful as a temporary conversion.
 * Serialize and deserialize must be paired; GC moving objects may also break it.
 */

// getSize returns the result.
func getSize(t any) int {
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return int(typ.Size())
}

// FromAny returns the result.
// 用 unsafe.Slice 构造以保持指针存活：经 reflect.SliceHeader 的 uintptr Data
// 字段绕行时 GC 不会把它当作存活引用，返回的 slice 可能悬垂。
func FromAny(s any) []byte {
	sizeOfStruct := getSize(s)
	data := (*reflectx.Eface)(unsafe.Pointer(&s)).Value
	return unsafe.Slice((*byte)(data), sizeOfStruct)
}
