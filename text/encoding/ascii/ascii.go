/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ascii

// Is c an ASCII lower-case letter?
func IsLower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func IsUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func IsLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// Is c an ASCII digit?
func IsDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func IsAllLower(s string) bool {
	for _, c := range s {
		if 'a' > c || c > 'z' {
			return false
		}
	}
	return true
}

func IsAllUpper(s string) bool {
	for _, c := range s {
		if 'A' > c || c > 'Z' {
			return false
		}
	}
	return true
}

func IsAllLetter(s string) bool {
	for _, c := range s {
		if c < 'A' || c > 'z' || (c > 'Z' && c < 'a') {
			return false
		}
	}
	return true
}

func EqualFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if Lower(s[i]) != Lower(t[i]) {
			return false
		}
	}
	return true
}

func Lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b ^ ' '
	}
	return b
}

func Upper(b byte) byte {
	if 'a' <= b && b <= 'z' {
		return b ^ ' '
	}
	return b
}
