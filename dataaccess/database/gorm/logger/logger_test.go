/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package logger

import (
	"gorm.io/gorm"
	"testing"
)

type Letter int

func (l Letter) String() string {
	switch l {
	case A:
		return "A"
	case B:
		return "B"
	default:
		return "C"
	}
}

const (
	A Letter = 1 + iota
	B
	C
)

type Foo struct {
	Id     int
	Letter Letter
}

func TestLogger(t *testing.T) {
	db := gorm.DB{}
	var foos []Foo
	db.Where("letter=?", A).Find(&foos)
}
