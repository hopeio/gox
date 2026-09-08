/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strings

import "unsafe"

//go:nosplit
//
// ToBytes returns a []byte that aliases the string's underlying read-only
// memory. The returned slice MUST NOT be written to: string memory is immutable,
// and mutating it will cause a segmentation fault. Use it only for read-only
// operations (e.g. passing a string to an API that requires []byte without
// copying). If you need a writable copy, use []byte(s) instead.
func ToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

//go:nosplit
func FromBytes(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
