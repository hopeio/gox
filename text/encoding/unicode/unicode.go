/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package unicode

import (
	"slices"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	stringsx "github.com/hopeio/gox/strings"
)

// [。；，：“”（）、？《》]
var HanPunctuation = []rune{
	'\u3002', '\uff1b', '\uff0c', '\uff1a', '\u201c', '\u201d', '\uff08', '\uff09', '\u3001', '\uff1f', '\u300a', '\u300b',
}

func HasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || slices.Contains(HanPunctuation, r) {
			return true
		}
	}
	return false
}

// Getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func Getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

func ToUtf8(s []byte) string {
	if len(s) < 6 {
		return stringsx.FromBytes(s)
	}
	b := make([]byte, len(s)+2*utf8.UTFMax)
	begin, bbegin := 0, 0
	for i := 0; i+6 <= len(s); {
		if s[i] == '\\' && s[i+1] == 'u' {
			bbegin += copy(b[bbegin:], s[begin:i])
			rr := Getu4(s[i:])
			if rr < 0 {
				return stringsx.FromBytes(s)
			}
			i += 6
			if utf16.IsSurrogate(rr) {
				rr1 := Getu4(s[i:])
				if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
					// A valid pair; consume.
					i += 6
					bbegin += utf8.EncodeRune(b[bbegin:], dec)
					break
				}
				// Invalid surrogate; fall back to replacement rune.
				rr = unicode.ReplacementChar
			}
			begin = i
			bbegin += utf8.EncodeRune(b[bbegin:], rr)
		} else {
			i++
		}
	}
	bbegin += copy(b[bbegin:], s[begin:])
	return stringsx.FromBytes(b[:bbegin])
}

func ToLowerFirst(s string) string {
	if len(s) > 0 {
		return string(unicode.ToLower(rune(s[0]))) + s[1:]
	}
	return ""
}
