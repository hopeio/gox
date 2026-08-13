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

// Int64 returns the result.
func Int64(b []byte) int64 {
	return int64(b[7]) | int64(b[6])<<8 | int64(b[5])<<16 | int64(b[4])<<24 |
		int64(b[3])<<32 | int64(b[2])<<40 | int64(b[1])<<48 | int64(b[0])<<56
}

// FromInt64 returns the result.
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

// Integer decodes b using the machine's NATIVE endianness (zero-copy cast).
// Only pair it with FromInteger within the same process; for cross-machine or
// persisted data use the big-endian Int64/FromInt64 instead.
// Panics if b is shorter than the size of T, like encoding/binary does.
func Integer[T constraints.Integer](b []byte) T {
	var v T
	if uintptr(len(b)) < unsafe.Sizeof(v) {
		panic("binary: buffer too small")
	}
	return *(*T)(unsafe.Pointer(unsafe.SliceData(b)))
}

// FromInteger encodes v using the machine's NATIVE endianness. See Integer.
func FromInteger[T constraints.Integer](v T) []byte {
	b := make([]byte, unsafe.Sizeof(v))
	*(*T)(unsafe.Pointer(unsafe.SliceData(b))) = v
	return b
}

// Int decodes a native-endian int. See Integer.
func Int(b []byte) int {
	return Integer[int](b)
}

// FromInt encodes a native-endian int. See Integer.
func FromInt(i int) []byte {
	return FromInteger(i)
}

// Uint decodes a native-endian uint64. See Integer.
func Uint(b []byte) uint64 {
	return Integer[uint64](b)
}

// FromUint encodes a native-endian uint64. See Integer.
func FromUint(i uint64) []byte {
	return FromInteger(i)
}
