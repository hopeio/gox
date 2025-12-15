/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package excel

type ColumnNumber int

const (
	A ColumnNumber = iota
	B
	C
	D
	E
	F
	G
	H
	I
	J
	K
	L
	M
	N
	O
	P
	Q
	R
	S
	T
	U
	V
	W
	X
	Y
	Z
	AA
	AB
	AC
)

// 只拓展到两位列ZZ
func (c ColumnNumber) Sting() string {
	if c < 26 {
		return string(rune(c + 'A'))
	}

	return (c/26 - 1).Sting() + (c % 26).Sting()
}

var ColumnLetter = [...]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "Value", "W", "X", "Y", "Z", "AA", "AB", "AC"}
