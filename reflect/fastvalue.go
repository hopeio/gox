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

package reflect

import (
	"reflect"
	"unsafe"
)

var (
	reflectRtypeItab = findReflectRtypeItab()
)

// Type.KindFlags const
const (
	F_direct    = 1 << 5
	F_kind_mask = (1 << 5) - 1
)

// Type.Flags const
const (
	tflagUncommon      uint8 = 1 << 0
	tflagExtraStar     uint8 = 1 << 1
	tflagNamed         uint8 = 1 << 2
	tflagRegularMemory uint8 = 1 << 3
)

// IsNamed reports whether the condition holds.
func (self *Type) IsNamed() bool {
	return (self.Flags & tflagNamed) != 0
}

// Kind returns the result.
func (self *Type) Kind() reflect.Kind {
	return reflect.Kind(self.KindFlags & F_kind_mask)
}

// Pack performs the operation.
func (self *Type) Pack() (t reflect.Type) {
	(*Iface)(unsafe.Pointer(&t)).Itab = reflectRtypeItab
	(*Iface)(unsafe.Pointer(&t)).Value = unsafe.Pointer(self)
	return
}

// String returns the string representation.
func (self *Type) String() string {
	return self.Pack().String()
}

// Indirect reports whether the condition holds.
func (self *Type) Indirect() bool {
	return self.KindFlags&F_direct == 0
}

type Map struct {
	Count      int
	Flags      uint8
	B          uint8
	Overflow   uint16
	Hash0      uint32
	Buckets    unsafe.Pointer
	OldBuckets unsafe.Pointer
	Evacuate   uintptr
	Extra      unsafe.Pointer
}

type MapIterator struct {
	K           unsafe.Pointer
	V           unsafe.Pointer
	T           *MapType
	H           *Map
	Buckets     unsafe.Pointer
	Bptr        *unsafe.Pointer
	Overflow    *[]unsafe.Pointer
	OldOverflow *[]unsafe.Pointer
	StartBucket uintptr
	Offset      uint8
	Wrapped     bool
	B           uint8
	I           uint8
	Bucket      uintptr
	CheckBucket uintptr
}

// Pack performs the operation.
func (self Eface) Pack() (v any) {
	*(*Eface)(unsafe.Pointer(&v)) = self
	return
}

// IndirectElem reports whether the condition holds.
func (self *MapType) IndirectElem() bool {
	return self.Flags&2 != 0
}

type Slice struct {
	Ptr unsafe.Pointer
	Len int
	Cap int
}

type String struct {
	Ptr unsafe.Pointer
	Len int
}

// PtrElem returns the result.
func PtrElem(t *Type) *Type {
	return (*PtrType)(unsafe.Pointer(t)).Elem
}

// ToMapType converts the value.
func ToMapType(t *Type) *MapType {
	return (*MapType)(unsafe.Pointer(t))
}

// IfaceType returns the result.
func IfaceType(t *Type) *InterfaceType {
	return (*InterfaceType)(unsafe.Pointer(t))
}

// UnpackType returns the result.
func UnpackType(t reflect.Type) *Type {
	return (*Type)((*Iface)(unsafe.Pointer(&t)).Value)
}

// UnpackEface returns the result.
func UnpackEface(v interface{}) Eface {
	return *(*Eface)(unsafe.Pointer(&v))
}

// UnpackIface returns the result.
func UnpackIface(v interface{}) Iface {
	return *(*Iface)(unsafe.Pointer(&v))
}

// findReflectRtypeItab returns the result.
func findReflectRtypeItab() *Itab {
	v := reflect.TypeOf(struct{}{})
	return (*Iface)(unsafe.Pointer(&v)).Itab
}

// AssertI2I performs the operation.
func AssertI2I(t *Type, i Iface) (r Iface) {
	inter := IfaceType(t)
	tab := i.Itab
	if tab == nil {
		return
	}
	if tab.Inter != inter {
		tab = GetItab(inter, tab.Type, true)
		if tab == nil {
			return
		}
	}
	r.Itab = tab
	r.Value = i.Value
	return
}

//go:noescape
//go:linkname GetItab runtime.getitab
func GetItab(inter *InterfaceType, typ *Type, canfail bool) *Itab

// GetFuncPC returns the value.
func GetFuncPC(fn any) uintptr {
	ft := UnpackEface(fn)
	if ft.Type.Kind() != reflect.Func {
		panic("not a function")
	}
	return *(*uintptr)(ft.Value)
}

// FuncAddr returns the result.
func FuncAddr(f any) unsafe.Pointer {
	if vv := UnpackEface(f); vv.Type.Kind() != reflect.Func {
		panic("f is not a function")
	} else {
		return *(*unsafe.Pointer)(vv.Value)
	}
}

// BytesFrom performs the operation.
func BytesFrom(p unsafe.Pointer, n int, c int) (r []byte) {
	(*Slice)(unsafe.Pointer(&r)).Ptr = p
	(*Slice)(unsafe.Pointer(&r)).Len = n
	(*Slice)(unsafe.Pointer(&r)).Cap = c
	return
}
