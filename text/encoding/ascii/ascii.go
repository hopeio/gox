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

// IsUpper reports whether the condition holds.
func IsUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

// IsLetter reports whether the condition holds.
func IsLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// Is c an ASCII digit?
func IsDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// IsAllLower reports whether the condition holds.
func IsAllLower(s string) bool {
	for _, c := range s {
		if 'a' > c || c > 'z' {
			return false
		}
	}
	return true
}

// IsAllUpper reports whether the condition holds.
func IsAllUpper(s string) bool {
	for _, c := range s {
		if 'A' > c || c > 'Z' {
			return false
		}
	}
	return true
}

// IsAllLetter reports whether the condition holds.
func IsAllLetter(s string) bool {
	for _, c := range s {
		if c < 'A' || c > 'z' || (c > 'Z' && c < 'a') {
			return false
		}
	}
	return true
}

// EqualFold reports whether the condition holds.
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

// Lower returns the result.
func Lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b ^ ' '
	}
	return b
}

// Upper returns the result.
func Upper(b byte) byte {
	if 'a' <= b && b <= 'z' {
		return b ^ ' '
	}
	return b
}
