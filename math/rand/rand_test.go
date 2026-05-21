/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package rand

import (
	"testing"
	"unicode"
)

func TestIntn(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := Intn(5, 10)
		if v < 5 || v >= 10 {
			t.Errorf("Intn(5,10) = %d, want [5,10)", v)
		}
	}
}

func TestChinese(t *testing.T) {
	s := Chinese()
	if len(s) == 0 {
		t.Error("Chinese() returned empty string")
	}
	for _, r := range s {
		if r < 19968 || r > 19968+500 {
			t.Errorf("Chinese() rune %d out of expected range", r)
		}
	}
}

func TestChineseChar(t *testing.T) {
	r := ChineseChar()
	if r < 19968 || r > 19968+500 {
		t.Errorf("ChineseChar() = %d, want in [19968, 20468]", r)
	}
}

func TestEnglish(t *testing.T) {
	s := English()
	if len(s) != 5 {
		t.Errorf("English() length = %d, want 5", len(s))
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			t.Errorf("English() char %q not in expected range", c)
		}
	}
}

func TestRandomCode(t *testing.T) {
	code := RandomCode(6)
	if len(code) != 6 {
		t.Errorf("RandomCode(6) length = %d, want 6", len(code))
	}
}

func TestRandomChars(t *testing.T) {
	s := RandomChars(10)
	if len(s) != 10 {
		t.Errorf("RandomChars(10) length = %d, want 10", len(s))
	}
	for _, c := range s {
		if !unicode.IsLetter(rune(c)) {
			t.Errorf("RandomChars() char %q not a letter", c)
		}
	}
}

func TestRandomNumber(t *testing.T) {
	s := RandomNumber(8)
	if len(s) != 8 {
		t.Errorf("RandomNumber(8) length = %d, want 8", len(s))
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Errorf("RandomNumber() char %q not a digit", c)
		}
	}
}
