/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package constraints

import (
	"encoding"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type Callback[T any] interface {
	~func() | ~func() error | ~func(T) | ~func(T) error
}

type Key interface {
	constraints.Integer | ~string | ~[8]byte | ~[16]byte | ~[32]byte | constraints.Float //| ~[]byte
}

type ID = Key

type Basic interface {
	Number | ~bool
}

type Ordered constraints.Ordered

type Enum interface {
	constraints.Integer
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}
