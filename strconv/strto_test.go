/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strconv

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestSizeof(t *testing.T) {
	t.Log(unsafe.Sizeof(1))
}

func TestStringConvertBasicFor(t *testing.T) {

	t.Run("int8", func(t *testing.T) {
		got, err := ParseFor[int8]("123")
		assert.Nil(t, err)
		assert.Equal(t, int8(123), got)
	})
	t.Run("int", func(t *testing.T) {
		got, err := ParseFor[int]("123456789")
		assert.Nil(t, err)
		assert.Equal(t, 123456789, got)
	})
	t.Run("uint", func(t *testing.T) {
		got, err := ParseFor[uint]("123456789")
		assert.Nil(t, err)
		assert.Equal(t, uint(123456789), got)
	})
	t.Run("bool", func(t *testing.T) {
		got, err := ParseFor[bool]("1")
		assert.Nil(t, err)
		assert.Equal(t, true, got)
	})
	t.Run("float32", func(t *testing.T) {
		got, err := ParseFor[float32]("1.23")
		assert.Nil(t, err)
		assert.Equal(t, float32(1.23), got)
	})
}
