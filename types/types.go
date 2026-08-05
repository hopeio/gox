/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	constraintsi "github.com/hopeio/gox/types/constraints"
)

type String string

// Key returns the underlying value.
func (s String) Key() string {
	return string(s)
}

type Int int

// Key returns the underlying value.
func (s Int) Key() int {
	return int(s)
}

type Basic struct {
}

type ID[T constraintsi.ID] struct {
	Id T `json:"id"`
}

// Key returns the wrapped ID value.
func (s ID[KEY]) Key() KEY {
	return s.Id
}
