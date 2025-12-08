/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binary

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

func Int64(b []byte) int64 {
	return int64(b[7]) | int64(b[6])<<8 | int64(b[5])<<16 | int64(b[4])<<24 |
		int64(b[3])<<32 | int64(b[2])<<40 | int64(b[1])<<48 | int64(b[0])<<56
}

func FromInt64(i int64) []byte {
	return []byte{
		byte(i >> 56),
		byte(i >> 48),
		byte(i >> 40),
		byte(i >> 32),
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i),
	}
}

func Integer[T constraints.Integer](b []byte) T {
	return *(*T)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b))))
}

func FromInteger[T constraints.Integer](v T) []byte {
	byteNum := unsafe.Sizeof(v)
	b := make([]byte, byteNum)
	*(*T)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b)))) = v
	return b
}

func Int(b []byte) int {
	return *(*int)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b))))
}

func FromInt(i int) []byte {
	b := make([]byte, 8)
	*(*int)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b)))) = i
	return b
}

// 比标准库慢很多,10倍左右，string和bytes互转只是节省复制内存，unsafe操作有很多检测
// binary.LittleEndian.Uint64(b)
func Uint(b []byte) uint64 {
	return *(*uint64)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b))))
}

// binary.LittleEndian.PutUint64(b)
func FromUint(i uint64) []byte {
	b := make([]byte, 8)
	*(*uint64)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b)))) = i
	return b
}
