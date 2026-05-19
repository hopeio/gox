package idgen

import (
	"testing"
)

func TestID_IsValid(t *testing.T) {
	// All zeros
	id := ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if id.IsValid() {
		t.Error("All-zero ID should not be valid")
	}
	// Non-zero
	id2 := ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	if !id2.IsValid() {
		t.Error("Non-zero ID should be valid")
	}
}

func TestID_String(t *testing.T) {
	id := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	s := id.String()
	if len(s) != 32 {
		t.Errorf("String() len = %d, want 32", len(s))
	}
}

func TestID_Hex(t *testing.T) {
	id := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	if id.Hex() != id.String() {
		t.Error("Hex() should equal String()")
	}
}

func TestID_Bytes(t *testing.T) {
	id := ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	b := id.Bytes()
	if len(b) != 16 {
		t.Errorf("Bytes() len = %d, want 16", len(b))
	}
}

func TestID_Base58(t *testing.T) {
	id := UniqueID()
	s := id.Base58()
	if len(s) == 0 {
		t.Error("Base58() returned empty string")
	}
}

func TestID_Base62(t *testing.T) {
	id := UniqueID()
	s := id.Base62()
	if len(s) == 0 {
		t.Error("Base62() returned empty string")
	}
}

func TestID_Base64(t *testing.T) {
	id := UniqueID()
	s := id.Base64()
	if len(s) == 0 {
		t.Error("Base64() returned empty string")
	}
}

func TestID_Base32(t *testing.T) {
	id := UniqueID()
	s := id.Base32()
	if len(s) == 0 {
		t.Error("Base32() returned empty string")
	}
}

func TestUniqueID(t *testing.T) {
	id1 := UniqueID()
	id2 := UniqueID()
	if !id1.IsValid() {
		t.Error("UniqueID() produced invalid ID")
	}
	// IDs should be unique (with very high probability)
	same := true
	for i := range id1 {
		if id1[i] != id2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("Two UniqueID() calls produced the same ID")
	}
}

func TestNewOrderedID(t *testing.T) {
	id1 := NewOrderedID()
	id2 := NewOrderedID()
	if id2 <= id1 {
		t.Errorf("OrderedIDs should be increasing: %d <= %d", id2, id1)
	}
}

func TestNewOrderedIDGenerator(t *testing.T) {
	gen := NewOrderedIDGenerator(100)
	if gen() != 101 {
		t.Errorf("NewOrderedIDGenerator(100)() = %d, want 101", gen())
	}
	if gen() != 102 {
		t.Errorf("second call = %d, want 102", gen())
	}
}

func TestID_MarshalJSON(t *testing.T) {
	id := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	data, err := id.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty")
	}
}
