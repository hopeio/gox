/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package encoding

import "io"

type Codec interface {
	Unmarshaler
	Marshaler
}

type Decoder interface {
	Decode(io.Reader, any) error
}

type Encoder interface {
	Encode(io.Writer, any) error
}

type Unmarshaler interface {
	Unmarshal([]byte, any) error
}

type Marshaler interface {
	Marshal(v any) ([]byte, error)
}
