/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	constraintsi "github.com/hopeio/gox/types/constraints"
	"golang.org/x/exp/constraints"
)

// CastSigned converts the value.
func CastSigned[T, V constraints.Signed](v V) T {
	return T(v)
}

// CastFloat converts the value.
func CastFloat[T, V constraints.Float](v V) T {
	return T(v)
}

// CastUnsigned converts the value.
func CastUnsigned[T, V constraints.Unsigned](v V) T {
	return T(v)
}

// CastInteger converts the value.
func CastInteger[T, V constraints.Integer](v V) T {
	return T(v)
}

// CastNumber converts the value.
func CastNumber[T, V constraintsi.Number](v V) T {
	return T(v)
}
