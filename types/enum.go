/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	constraintsi "github.com/hopeio/gox/types/constraints"
)

type Enum[T constraintsi.Enum] int32

// MarshalText ...
func (t Enum[T]) MarshalText() ([]byte, error) {
	return T(t).MarshalText()
}

// UnmarshalText ...
func (t *Enum[T]) UnmarshalText(data []byte) error {
	var tt T
	err := tt.UnmarshalText(data)
	if err != nil {
		return err
	}
	*t = Enum[T](tt)
	return nil
}

// UnmarshalJSON ...
func (t *Enum[T]) UnmarshalJSON(data []byte) error {
	data = data[1 : len(data)-1]
	return t.UnmarshalText(data)
}

// MarshalJSON ...
func (t Enum[T]) MarshalJSON() ([]byte, error) {
	text, err := t.MarshalText()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(text)+2)
	buf[0] = '"'
	buf = append(buf, text...)
	buf[len(buf)-1] = '"'
	return buf, nil
}
