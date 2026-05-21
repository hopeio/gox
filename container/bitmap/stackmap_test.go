/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bitmap

import (
	"testing"
	"unsafe"
)

func unsafePointer(p *byte) unsafe.Pointer {
	return unsafe.Pointer(p)
}

func TestBitmap_Append(t *testing.T) {
	var b Bitmap
	b.Append(1)
	b.Append(0)
	b.Append(1)
	b.Append(1)

	if b.N != 4 {
		t.Errorf("N = %d, want 4", b.N)
	}
	if len(b.B) == 0 {
		t.Error("B should not be empty after Append")
	}
}

func TestBitmap_AppendMany(t *testing.T) {
	var b Bitmap
	b.AppendMany(5, 1)
	if b.N != 5 {
		t.Errorf("N = %d, want 5", b.N)
	}

	b2 := Bitmap{}
	b2.AppendMany(3, 0)
	if b2.N != 3 {
		t.Errorf("N = %d, want 3", b2.N)
	}
}

func TestBitmap_Set(t *testing.T) {
	var b Bitmap
	b.Append(0)
	b.Append(1)
	b.Append(0)

	b.Set(1, 0)
	// Verify set works without panic
	b.Set(0, 1)
}

func TestBitmap_SetInvalidPosition(t *testing.T) {
	var b Bitmap
	b.Append(0)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Set() with invalid position should panic")
		}
	}()
	b.Set(5, 1)
}

func TestBitVec(t *testing.T) {
	// data = 0b10110100 = 0xB4 = 180
	// Bit(0) = (180 >> 0) & 1 = 0
	// Bit(2) = (180 >> 2) & 1 = 1
	// Bit(3) = (180 >> 3) & 1 = 0
	// Bit(4) = (180 >> 4) & 1 = 1
	data := []byte{0b10110100}
	bv := BitVec{N: 8, B: unsafePointer(&data[0])}

	if bv.Bit(0) != 0 {
		t.Errorf("Bit(0) = %d, want 0", bv.Bit(0))
	}
	if bv.Bit(2) != 1 {
		t.Errorf("Bit(2) = %d, want 1", bv.Bit(2))
	}
	if bv.Bit(3) != 0 {
		t.Errorf("Bit(3) = %d, want 0", bv.Bit(3))
	}
	if bv.Bit(4) != 1 {
		t.Errorf("Bit(4) = %d, want 1", bv.Bit(4))
	}
}

func TestBitVec_String(t *testing.T) {
	data := []byte{0b1010}
	bv := BitVec{N: 4, B: unsafePointer(&data[0])}
	s := bv.String()
	if len(s) == 0 {
		t.Error("String() should not be empty")
	}
}
