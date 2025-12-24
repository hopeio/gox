/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

/*
 * Copyright 2021 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package strings

import (
	"unsafe"

	reflectx "github.com/hopeio/gox/reflect"
)

//go:nosplit
func Bytes2Str(v []byte) (s string) {
	(*reflectx.String)(unsafe.Pointer(&s)).Len = (*reflectx.Slice)(unsafe.Pointer(&v)).Len
	(*reflectx.String)(unsafe.Pointer(&s)).Ptr = (*reflectx.Slice)(unsafe.Pointer(&v)).Ptr
	return
}

//go:nosplit
func Str2Bytes(s string) (v []byte) {
	(*reflectx.Slice)(unsafe.Pointer(&v)).Cap = (*reflectx.String)(unsafe.Pointer(&s)).Len
	(*reflectx.Slice)(unsafe.Pointer(&v)).Len = (*reflectx.String)(unsafe.Pointer(&s)).Len
	(*reflectx.Slice)(unsafe.Pointer(&v)).Ptr = (*reflectx.String)(unsafe.Pointer(&s)).Ptr
	return
}

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
