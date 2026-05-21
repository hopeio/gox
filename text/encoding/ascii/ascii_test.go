/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ascii

import "testing"

func TestIsLower(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'a', true},
		{'z', true},
		{'m', true},
		{'A', false},
		{'Z', false},
		{'0', false},
		{'@', false},
	}
	for _, tt := range tests {
		if got := IsLower(tt.input); got != tt.want {
			t.Errorf("IsLower(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsUpper(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'A', true},
		{'Z', true},
		{'M', true},
		{'a', false},
		{'z', false},
		{'0', false},
	}
	for _, tt := range tests {
		if got := IsUpper(tt.input); got != tt.want {
			t.Errorf("IsUpper(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'a', true},
		{'Z', true},
		{'0', false},
		{' ', false},
	}
	for _, tt := range tests {
		if got := IsLetter(tt.input); got != tt.want {
			t.Errorf("IsLetter(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'0', true},
		{'9', true},
		{'5', true},
		{'a', false},
		{'A', false},
		{'/', false},
		{':', false},
	}
	for _, tt := range tests {
		if got := IsDigit(tt.input); got != tt.want {
			t.Errorf("IsDigit(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsAllLower(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc", true},
		{"hello", true},
		{"ABC", false},
		{"Abc", false},
		{"ab1", false},
		{"", true},
	}
	for _, tt := range tests {
		if got := IsAllLower(tt.input); got != tt.want {
			t.Errorf("IsAllLower(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsAllUpper(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"ABC", true},
		{"HELLO", true},
		{"abc", false},
		{"AbC", false},
		{"AB1", false},
		{"", true},
	}
	for _, tt := range tests {
		if got := IsAllUpper(tt.input); got != tt.want {
			t.Errorf("IsAllUpper(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsAllLetter(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc", true},
		{"ABC", true},
		{"AbC", true},
		{"ab1", false},
		{"", true},
	}
	for _, tt := range tests {
		if got := IsAllLetter(tt.input); got != tt.want {
			t.Errorf("IsAllLetter(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'M', 'm'},
		{'a', 'a'},
		{'0', '0'},
		{'@', '@'},
	}
	for _, tt := range tests {
		if got := Lower(tt.input); got != tt.want {
			t.Errorf("Lower(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestUpper(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
	}{
		{'a', 'A'},
		{'z', 'Z'},
		{'m', 'M'},
		{'A', 'A'},
		{'0', '0'},
	}
	for _, tt := range tests {
		if got := Upper(tt.input); got != tt.want {
			t.Errorf("Upper(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEqualFold(t *testing.T) {
	tests := []struct {
		s, t string
		want bool
	}{
		{"abc", "ABC", true},
		{"Hello", "HELLO", true},
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "ab", false},
		{"", "", true},
		{"Go", "go", true},
	}
	for _, tt := range tests {
		if got := EqualFold(tt.s, tt.t); got != tt.want {
			t.Errorf("EqualFold(%q, %q) = %v, want %v", tt.s, tt.t, got, tt.want)
		}
	}
}
