/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sync

import (
	"math"
	"sync/atomic"
	"unsafe"
)

// AddFloat64 add delta to given address atomically
func AddFloat64(addr *float64, delta float64) (new float64) {
	var old float64
	for {
		old = math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(addr))))
		new = old + delta
		if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(addr)),
			math.Float64bits(old), math.Float64bits(new)) {
			break
		}
	}
	return
}

// AddFloat32 ...
func AddFloat32(addr *float32, delta float32) (new float32) {
	var old float32
	for {
		old = math.Float32frombits(atomic.LoadUint32((*uint32)(unsafe.Pointer(addr))))
		new = old + delta
		if atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(addr)),
			math.Float32bits(old), math.Float32bits(new)) {
			break
		}
	}
	return
}

// SubUint32 ...
func SubUint32(addr *uint32, delta uint32) (new uint32) {
	return atomic.AddUint32(addr, ^uint32(delta-1))
}

// SubUint64 ...
func SubUint64(addr *uint64, delta uint64) (new uint64) {
	return atomic.AddUint64(addr, ^uint64(delta-1))
}
