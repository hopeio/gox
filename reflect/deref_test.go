/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reflect

import (
	"reflect"
	"testing"
)

func TestDerefValue_NilAndValid(t *testing.T) {
	var p *int
	if DerefValue(reflect.ValueOf(p)).IsValid() {
		t.Fatal("nil pointer should be invalid")
	}
	var a any
	if DerefValue(reflect.ValueOf(&a).Elem()).IsValid() {
		t.Fatal("nil interface should be invalid")
	}
	var a2 any = (*int)(nil)
	if DerefValue(reflect.ValueOf(&a2).Elem()).IsValid() {
		t.Fatal("interface holding nil pointer should be invalid")
	}
	a3 := any(1)
	dv := DerefValue(reflect.ValueOf(&a3).Elem())
	if !dv.IsValid() || dv.Int() != 1 {
		t.Fatalf("got %#v", dv)
	}
	x := 9
	pp := &x
	ppp := &pp
	dv2 := DerefValue(reflect.ValueOf(ppp))
	if !dv2.IsValid() || dv2.Int() != 9 {
		t.Fatalf("multi ptr got %#v", dv2)
	}
}

func TestDerefType(t *testing.T) {
	typ := reflect.TypeOf((**[]int)(nil))
	got := DerefType(typ)
	if got.Kind() != reflect.Int {
		t.Fatalf("got %v", got)
	}
}
