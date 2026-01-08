//go:build go1.20

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strings

import "unsafe"

//go:nosplit
func ToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

//go:nosplit
func FromBytes(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
