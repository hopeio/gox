/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	"github.com/hopeio/gox/reflect"
	"unsafe"
)

// GrowSlice returns the result.
func GrowSlice(et *reflect.Type, old reflect.Slice, cap int) reflect.Slice {
	s := growslice(old.Ptr, cap, old.Cap, cap-old.Len, et)
	s.Len = old.Len
	return s
}

//go:linkname growslice runtime.growslice
//goland:noinspection GoUnusedParameter
func growslice(oldPtr unsafe.Pointer, newLen, oldCap, num int, et *reflect.Type) reflect.Slice
