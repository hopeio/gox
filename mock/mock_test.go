/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mock

import (
	"encoding/json"
	"log"
	"testing"
	"unicode/utf8"
)

type Bar1 struct {
	Field1 int
	Field2 string `mock:"example:'1',type:'\\w'"`
	Field3 *Bar2
}

type Bar2 struct {
	Field1 int
	Field2 string
	Field3 []*Bar3
}

type Bar3 struct {
	Field1 float64
	Field2 string
	Field3 map[string]int
}

func TestMock(t *testing.T) {
	var bar Bar1
	Mock(&bar)
	data, err := json.Marshal(&bar)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(data))
}

func TestUnicode(t *testing.T) {
	r1, i1 := utf8.DecodeRuneInString("\u4e00")
	log.Println(r1, i1)
	r2, i2 := utf8.DecodeRuneInString("\u9fa5")
	log.Println(r2, i2)
	log.Println(r2 - r1)
}
