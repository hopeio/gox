/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package hash

import (
	"reflect"
	"testing"
	"time"
)

type Foo struct {
	Time time.Time
	Bar
}

type Bar struct {
	Int int
}

func TestMarshal(t *testing.T) {
	e := new(encodeState)
	u := &Foo{Time: time.Now(), Bar: Bar{Int: 1}}
	e.encode("", reflect.ValueOf(u))
	for i := 0; i < len(e.strings); i += 2 {
		t.Log(e.strings[i], e.strings[i+1])
	}
}
