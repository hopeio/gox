/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reflect

import (
	"reflect"
	"unsafe"
)

type Iface struct {
	Itab  *Itab
	Value unsafe.Pointer
}

type Eface struct {
	Type  *Type
	Value unsafe.Pointer
}

type Itab struct {
	Inter  *InterfaceType
	Type   *Type
	Hash   uint32 // copy of _type.hash. Used for type switches.
	Unused [4]byte
	Fun    [1]uintptr // variable sized
}

// interfaceType represents an interface type.
type InterfaceType struct {
	Type
	PkgPath *byte     // import path
	Methods []IMethod // sorted by hash
}

// Type must be kept in sync with ../runtime/type.go:/^type._type.
type Type struct {
	Size       uintptr
	Ptrdata    uintptr // number of bytes in the type that can contain pointers
	Hash       uint32  // hash of type; avoids computation in hash tables
	Flags      uint8   // extra type information flags
	Align      uint8   // alignment of variable with this type
	FieldAlign uint8   // alignment of struct field with this type
	KindFlags  uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	Gcdata    *byte   // garbage collection data
	Str       NameOff // string form
	PtrToThis TypeOff // type for pointer to this type, may be zero
}

// IMethod represents a method on an interface type
type IMethod struct {
	Name NameOff // name of method
	Typ  TypeOff // .(*FuncType) underneath
}

type NameOff int32 // offset to a name
type TypeOff int32 // offset to an *Rtype
type TextOff int32 // offset from top of text section

// TypeByOff ...
func TypeByOff(section unsafe.Pointer, off int32) *Type {
	return (*Type)(add(section, uintptr(off)))
}

// add ...
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

// ArrayType represents a fixed array type.
type ArrayType struct {
	Type
	Elem  *Type // array element type
	Slice *Type // Slice type
	Len   uintptr
}

// ChanType represents a channel type.
type ChanType struct {
	Type
	Elem *Type   // channel element type
	Dir  uintptr // channel direction (ChanDir)
}

// FuncType represents a function type.
//
// A *Rtype for each in and out parameter is stored in an array that
// directly follows the funcType (and possibly its uncommonType). So
// a function type with one method, one input, and one output is:
//
//	struct {
//		funcType
//		uncommonType
//		[2]*Rtype    // [0] is in, [1] is out
//	}
type FuncType struct {
	Type
	InCount  uint16
	OutCount uint16 // top bit is set if last input parameter is ...
}

// MapType represents a map type.
type MapType struct {
	Type
	Key    *Type // map key type
	Elem   *Type // map element (value) type
	Bucket *Type // internal bucket structure
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	KeySize    uint8  // size of key slot
	ValueSize  uint8  // size of value slot
	BucketSize uint16 // size of bucket
	Flags      uint32
}

// PtrType represents a pointer type.
type PtrType struct {
	Type
	Elem *Type // pointer element (pointed at) type
}

// SliceType represents a Slice type.
type SliceType struct {
	Type
	Elem *Type // Slice element type
}

// StructType represents a struct type.
type StructType struct {
	Type
	PkgPath *byte
	Fields  []StructField // sorted by offset
}

// Struct field
type StructField struct {
	Name        *byte   // name is always non-empty
	Typ         *Type   // type of field
	OffsetEmbed uintptr // byte offset of field<<1 | isEmbedded
}

type FuncID uint8

type Func struct {
	Entry   uintptr // start pc
	Nameoff int32   // function name

	Args        int32  // in/out args size
	Deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	Pcsp      uint32
	Pcfile    uint32
	Pcln      uint32
	Npcdata   uint32
	CuOffset  uint32  // runtime.cutab offset of this function's CU
	FuncID    FuncID  // set for certain special runtime functions
	_         [2]byte // pad
	Nfuncdata uint8   // must be last
}

// Hash reports whether the condition holds.
func Hash(v any) uint32 {
	ia := *(*Iface)(unsafe.Pointer(&v))
	return ia.Itab.Hash
}

type Flag uintptr

type Value struct {
	Typ *Type
	Ptr unsafe.Pointer
	Flag
}

// RuntimeTypeID ...
func RuntimeTypeID(t reflect.Type) uintptr {
	return uintptrElem(uintptr(unsafe.Pointer(&t)) + PtrOffset)
}

// DerefType ...
func DerefType(typ reflect.Type) reflect.Type {
	for {
		kind := typ.Kind()
		if kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Map || kind == reflect.Chan || kind == reflect.Array {
			typ = typ.Elem()
		} else {
			return typ
		}
	}
}
