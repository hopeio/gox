/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package unsafe

import (
	"unsafe"
)

//go:nosplit
//goland:noinspection GoVetUnsafePointer
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func Cast[T1, T2 any](p *T2) *T1 {
	return (*T1)(unsafe.Pointer(p))
}

func CastSlice[T1, T2 any](s []T2) []T1 {
	return unsafe.Slice((*T1)(unsafe.Pointer(unsafe.SliceData(s))), len(s))
}

func Binary(p unsafe.Pointer, n int) (r []byte) {
	return unsafe.Slice((*byte)(p), n)
}

func Clear[T any](ptr *T) {
	clear(unsafe.Slice(ptr, 1))
}
