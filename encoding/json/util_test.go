/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package json

import (
	"log"
	"testing"
)

func TestUnquote(t *testing.T) {
	s := []byte(`"\u8bf7\u6c42\u8fc7\u4e8e\u9891\u7e41"`)
	log.Println(Unquote(s))
}
