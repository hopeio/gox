/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */
package datatypes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Foo struct {
	A int
	B string
}

func TestJSON(t *testing.T) {
	var jat Json[[]Foo]
	err := jat.Scan([]byte(`[{"A":1,"B":"1"},{"A":2,"B":"2"}]`))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, Json[[]Foo]{V: []Foo{{1, "1"}, {2, "2"}}}, jat)
}
