/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package encoding

type Decoder interface {
	Decode(v interface{}) (err error)
}

type Encoder interface {
	Encode(v interface{}) (err error)
}

type Unmarshaler interface {
	Unmarshal([]byte, any) error
}

type Marshaler interface {
	Marshal(v any) ([]byte, error)
}
