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

func TestDerefInterfaceType(t *testing.T) {
	var a any
	a = 1
	v := reflect.TypeOf(&a)
	t.Log(v.Kind())
	v1 := v.Elem()
	t.Log(v1.Kind())
	// interface 类型本身不能 Elem；动态类型用 Value 取。
	if v1.Kind() != reflect.Interface {
		t.Fatalf("want interface, got %v", v1.Kind())
	}
	iv := reflect.ValueOf(&a).Elem()
	dyn := iv.Elem().Type()
	t.Log(dyn.Kind(), dyn)
	if dyn.Kind() != reflect.Int {
		t.Fatalf("want int dynamic type, got %v", dyn)
	}
}
