/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binary

import (
	"encoding/binary"
	"testing"
)

func TestInt64To(t *testing.T) {
	b := []byte{0, 1, 0, 0, 0, 0, 0, 0}
	t.Log(Int64(b))
	t.Log(binary.BigEndian.Uint64(b))
	b = FromInt64(15)
	t.Log(Int64(b))
}

func TestIntToUint(t *testing.T) {
	var i = -1
	b := binary.LittleEndian.AppendUint64(nil, uint64(i))
	j := binary.LittleEndian.Uint64(b)
	t.Log(uint(j))
}

func TestUintTo(t *testing.T) {
	b := FromUint(1111)
	t.Log(Uint(b))
}

func TestIntToB(t *testing.T) {
	b := FromInteger(1111)
	t.Log(Integer[int](b))
	t.Log(Integer[int](b))
}

func FuzzIntTo(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Fuzz(func(t *testing.T, i int64) {
		b := FromInt64(i)
		j := Int64(b)
		if j != i {
			t.Fatal(j, i)
		}
	})
}

func BenchmarkIntFromBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bb := FromUint(15)
		Uint(bb)
	}
}

func BenchmarkBigEndianUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bb := make([]byte, 8)
		binary.BigEndian.PutUint64(bb, 15)
		binary.BigEndian.Uint64(bb)
	}
}
