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

type Foo struct {
	A int
	B string
	C *int
	D *map[int]any
	E *[]int
}
type Bar struct {
	Foo Foo
	C   string
}

func TestInitStruct(t *testing.T) {
	var f *Foo
	v := reflect.ValueOf(&f)
	InitValue(v)
	t.Log(*f)
	t.Log(*f.C)
	t.Log(*f.D)
	t.Log(*f.E)
	t.Log(v.Elem().Interface())
}
