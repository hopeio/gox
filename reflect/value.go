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

const (
	FlagKindWidth        = 5 // there are 27 kinds
	FlagKindMask    Flag = 1<<FlagKindWidth - 1
	FlagStickyRO    Flag = 1 << 5
	FlagEmbedRO     Flag = 1 << 6
	FlagIndir       Flag = 1 << 7
	FlagAddr        Flag = 1 << 8
	FlagMethod      Flag = 1 << 9
	FlagMethodShift      = 10
	FlagRO          Flag = FlagStickyRO | FlagEmbedRO
)

var (
	e         = Eface{Type: new(Type)}
	PtrOffset = func() uintptr {
		return unsafe.Offsetof(e.Value)
	}()
	KindOffset = func() uintptr { return unsafe.Offsetof(e.Type.KindFlags) }()
	ElemOffset = func() uintptr {
		return unsafe.Offsetof(new(PtrType).Elem)
	}()
	SliceDataOffset = func() uintptr {
		return unsafe.Offsetof(new(reflect.SliceHeader).Data)
	}()
)

// DereferenceValue dereference and unpack interface,
// get the underlying non-pointer and non-interface value.
// if v is nil will return invalid value, shuold use IsValid() to check.
func DerefValue(v reflect.Value) reflect.Value {
	for {
		kind := v.Kind()
		if kind == reflect.Ptr || kind == reflect.Interface {
			v = v.Elem()
			if !v.IsValid() {
				return v
			}
		} else {
			return v
		}
	}
}

// InitPtr returns the result.
func InitPtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if !v.IsValid() || v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

//go:nocheckptr
func ValueOf(v interface{}) Value {
	stdValue := reflect.ValueOf(v)
	return *(*Value)(unsafe.Pointer(&stdValue))
}

//go:nocheckptr
func ConvertValue(v reflect.Value) Value {
	return *(*Value)(unsafe.Pointer(&v))
}

//go:nocheckptr
func getFlag(typPtr uintptr) Flag {
	if unsafe.Pointer(typPtr) == nil {
		return 0
	}
	return *(*Flag)(unsafe.Pointer(typPtr + KindOffset))
}

//go:nocheckptr
func pointerElem(p unsafe.Pointer) unsafe.Pointer {
	return *(*unsafe.Pointer)(p)
}

// Pointer gets the pointer of i.
// NOTE:
//
//	*T and T, gets diffrent pointer
//
//go:nocheckptr
func (v Value) Pointer() uintptr {
	switch v.Kind() {
	case reflect.Invalid:
		return 0
	case reflect.Slice:
		return uintptrElem(uintptr(v.Ptr)) + SliceDataOffset
	default:
		return uintptr(v.Ptr)
	}
}

// Kind gets the reflect.Kind fastly.
func (v Value) Kind() reflect.Kind {
	return reflect.Kind(v.Flag & FlagKindMask)
}

//go:nocheckptr
func uintptrElem(ptr uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(ptr))
}

// todo
func rangeValue(v reflect.Value, callbacks [reflect.UnsafePointer]func(reflect.Value) reflect.Value) {
	callbacks[v.Kind()](v)
}

// IsInteger reports whether the condition holds.
func IsInteger(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Uintptr
}

// IsFloat reports whether the condition holds.
func IsFloat(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}

// IsNumber reports whether the condition holds.
func IsNumber(kind reflect.Kind) bool {
	return IsInteger(kind) || IsFloat(kind)
}
