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

// Cast converts p to *T1 without allocation.
func Cast[T1, T2 any](p *T2) *T1 {
	return (*T1)(unsafe.Pointer(p))
}

// CastSlice converts a slice of T2 into a slice of T1 without allocation.
func CastSlice[T1, T2 any](s []T2) []T1 {
	return unsafe.Slice((*T1)(unsafe.Pointer(unsafe.SliceData(s))), len(s))
}

// Binary creates a byte slice view over memory at p with length n.
func Binary(p unsafe.Pointer, n int) (r []byte) {
	return unsafe.Slice((*byte)(p), n)
}

// Clear sets the single element at ptr to its zero value.
func Clear[T any](ptr *T) {
	clear(unsafe.Slice(ptr, 1))
}
