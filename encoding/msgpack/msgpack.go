/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package msgpack

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

var handler = codec.MsgpackHandle{}

func Marshal(v any) ([]byte, error) {
	r := bytes.NewBuffer(nil)
	decoder := codec.NewEncoder(r, &handler)
	return r.Bytes(), decoder.Encode(v)
}
