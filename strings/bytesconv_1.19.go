//go:build !go1.20

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strings

import "unsafe"

// 这个方式好一点,新建一个结构体承载
//
//go:nosplit
func ToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

//go:nosplit
func FromBytes(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// 这个直接用slice类型强转,会丢失cap信息
func toBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
