/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package json

import (
	"bytes"
	"fmt"

	"github.com/hopeio/gox/strings"
	unicodex "github.com/hopeio/gox/unicode"

	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func Unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = strings.BytesToString(s)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := unicodex.Getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := unicodex.Getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

func DecodeInt(b []byte) (int, []byte, error) {
	idx := bytes.IndexByte(b, ',')
	if idx == -1 {
		idx = bytes.IndexByte(b, '}')
		if idx == -1 {
			return -1, b, fmt.Errorf("no comma")
		}
	}
	i, err := strconv.Atoi(string(b[:idx]))
	if err != nil {
		return -1, b, err
	}
	return i, b[idx+1:], nil
}

func DecodeFloat(b []byte) (float64, []byte, error) {
	idx := bytes.IndexByte(b, ',')
	if idx == -1 {
		idx = bytes.IndexByte(b, '}')
		if idx == -1 {
			return -1, b, fmt.Errorf("no comma")
		}
	}
	f, err := strconv.ParseFloat(string(b[:idx]), 64)
	if err != nil {
		return -1, b, err
	}
	return f, b[idx+1:], nil
}

func DecodeString(b []byte) (string, []byte, error) {
	idx := bytes.Index(b, []byte(`",`))
	if idx == -1 {
		idx = bytes.IndexByte(b, '}')
		if idx == -1 {
			return "", b, fmt.Errorf("no comma")
		}
	}
	// Opening '"'
	if len(b) < 1 {
		return "", b, fmt.Errorf("%d", len(b))
	}
	s := string(b[1:idx])
	return s, b[idx+2:], nil
}

func DecodeBool(b []byte) (bool, []byte, error) {
	idx := bytes.IndexByte(b, ',')
	if idx == -1 {
		idx = bytes.IndexByte(b, '}')
		if idx == -1 {
			return false, b, fmt.Errorf("no comma")
		}
	}
	bol, err := strconv.ParseBool(string(b[:idx]))
	if err != nil {
		return false, b, err
	}
	return bol, b[idx+1:], nil
}
