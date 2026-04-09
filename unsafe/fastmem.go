/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package unsafe

import (
	"unsafe"

	reflectx "github.com/hopeio/gox/reflect"
)

//go:nocheckptr
func IndexChar(src string, index int) unsafe.Pointer {
	return unsafe.Pointer(uintptr((*reflectx.String)(unsafe.Pointer(&src)).Ptr) + uintptr(index))
}

//go:nocheckptr
func IndexByte(ptr []byte, index int) unsafe.Pointer {
	return unsafe.Pointer(uintptr((*reflectx.Slice)(unsafe.Pointer(&ptr)).Ptr) + uintptr(index))
}

//go:nosplit
func UnsafePtr(s string) unsafe.Pointer {
	return (*reflectx.String)(unsafe.Pointer(&s)).Ptr
}

//go:nosplit
func FromUnsafePtr(p unsafe.Pointer, n int64) (s string) {
	(*reflectx.String)(unsafe.Pointer(&s)).Ptr = p
	(*reflectx.String)(unsafe.Pointer(&s)).Len = int(n)
	return
}
