/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strings

import (
	"fmt"
	"testing"
)

func TestStrconv(t *testing.T) {
	s := "test"
	b := StringToBytes(s)
	s2 := BytesToString(b)
	fmt.Println(b, s2)
}
