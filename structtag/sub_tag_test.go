/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package structtag

import (
	"reflect"
	"testing"
)

type Bar1 struct {
	Field1 int    `type:"\\w"`
	Field2 string `mock:"example:'1' type:'\\w' test:'\"'"` // 因为标准库的Tag.Get最后会Unquote，自定义的无需再去Unquote
}

func TestTag(t *testing.T) {
	var bar Bar1
	typ := reflect.TypeOf(bar)
	t.Log(typ.Field(0).Tag.Get("type"))
	t.Log(SubTagLookup(typ.Field(1).Tag.Get("mock"), "type"))
	t.Log(SubTagLookup(typ.Field(1).Tag.Get("mock"), "test"))
	t.Log(SubTagLookup(typ.Field(1).Tag.Get("mock"), "example"))
}
