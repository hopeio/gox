/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package md5_test

import (
	"testing"

	"github.com/hopeio/gox/crypto/md5"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		value string
		want  string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := md5.EncodeString(tt.value)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("EncodeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		value string
		want  []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := md5.Encode(tt.value)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
