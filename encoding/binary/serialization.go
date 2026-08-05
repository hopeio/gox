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
	size := reflect.TypeOf(t).Elem().Size()
	return (int)(size)
}

// FromAny returns the result.
func FromAny(s any) []byte {
	sizeOfStruct := getSize(s)
	var x reflect.SliceHeader
	x.Len = sizeOfStruct
	x.Cap = sizeOfStruct
	x.Data = uintptr((*reflectx.Eface)(unsafe.Pointer(&s)).Value)
	return *(*[]byte)(unsafe.Pointer(&x))
}
