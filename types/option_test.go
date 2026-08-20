/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	"strconv"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func itoa(v int) string { return strconv.Itoa(v) }

func TestOptionP(t *testing.T) {
	v := None[int]()
	t.Log(v.IsSome())
	t.Log(v.IsNone())
	data, err := v.MarshalJSON()
	t.Log(string(data), err)
	v.IfSome(func(value int) {
		t.Log(value)
	})
	v.IfNone(func() {
		t.Log("none")
	})
}

func TestOptionPtr_Map(t *testing.T) {
	// present: maps *int -> *string
	opt := SomePtr(ptr(42))
	mapped := opt.Map(func(v *int) *string { s := "v" + itoa(*v); return &s })
	if !mapped.IsSome() {
		t.Fatal("Map: IsSome = false, want true")
	}
	if got := *mapped.Unwrap(); got != "v42" {
		t.Fatalf("Map = %q, want v42", got)
	}

	// none: mapping is skipped
	none := NonePtr[int]()
	noneMapped := none.Map(func(v *int) *string { return new(string) })
	if noneMapped.IsSome() {
		t.Fatal("Map on None: IsSome = true, want false")
	}

	// package-level MapOptionPtr delegates to the method
	delegated := MapOptionPtr(opt, func(v *int) *string { s := "d" + itoa(*v); return &s })
	if !delegated.IsSome() || *delegated.Unwrap() != "d42" {
		t.Fatalf("MapOptionPtr = %v, want d42", delegated)
	}
}
