/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package bits

import "unsafe"

const (
	BaseNaN Num64 = 0x7FF8000000000001
)

type Num64 uint64

// Float64ToNum64 ...
func Float64ToNum64(f float64) Num64 {
	return *(*Num64)(unsafe.Pointer(&f))
}

// Uint64ToNum64 ...
func Uint64ToNum64(f int64) Num64 {
	return *(*Num64)(unsafe.Pointer(&f))
}

// Int64ToNum64 ...
func Int64ToNum64(f uint64) Num64 {
	return Num64(f)
}

// At ...
func (b Num64) At(leftIdx int) bool {
	return b&(1<<(63-leftIdx)) != 0
}

// SetAt ...
func (b Num64) SetAt(leftIdx int, v bool) Num64 {
	if v {
		return b | (1 << (63 - leftIdx))
	}
	return b &^ (1 << (63 - leftIdx))
}

// SetRange ...
func (b Num64) SetRange(l, r int, v bool) Num64 {
	for ; l <= r; l++ {
		b = b.SetAt(l, v)
	}
	return b
}

// RangeInt ...
func (b Num64) RangeInt(leftIdx, rightIdx int) int64 {
	if b.At(leftIdx) {
		return -int64(^(b<<(leftIdx+1)>>(63-rightIdx+leftIdx+1))&1<<(rightIdx-leftIdx) + 1)
	}
	return int64(b << (leftIdx + 1) >> (63 - rightIdx + leftIdx + 1))
}

// RangeUint ...
func (b Num64) RangeUint(leftIdx, rightIdx int) uint64 {
	return uint64(b << (leftIdx + 1) >> (63 - rightIdx + leftIdx + 1))
}

// Float ...
func (b Num64) Float() float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// Uint ...
func (b Num64) Uint() uint64 {
	return uint64(b)
}

// Int ...
func (b Num64) Int() int64 {
	return *(*int64)(unsafe.Pointer(&b))
}

// Float64ToInt64 ...
func Float64ToInt64(f float64) int64 {
	return *(*int64)(unsafe.Pointer(&f))
}

// Int64ToFloat64 ...
func Int64ToFloat64(f int64) float64 {
	return *(*float64)(unsafe.Pointer(&f))
}

// Float64ToUint64 ...
func Float64ToUint64(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// Uint64ToFloat64 ...
func Uint64ToFloat64(f uint64) float64 {
	return *(*float64)(unsafe.Pointer(&f))
}

// Int64ToUint64 ...
func Int64ToUint64(f int64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// Uint64ToInt64 ...
func Uint64ToInt64(f uint64) int64 {
	return *(*int64)(unsafe.Pointer(&f))
}

type Uint8 uint8

// Num8 ...
func (b Num64) Num8(leftIdx int) bool {
	return b&(1<<(7-leftIdx)) != 0
}

// SetAt ...
func (b Uint8) SetAt(leftIdx int, v bool) Num64 {
	if v {
		return Num64(b | (1 << (7 - leftIdx)))
	}
	return Num64(b &^ (1 << (7 - leftIdx)))
}
