package base58

import (
	"testing"
)

func TestEncodeToString_Empty(t *testing.T) {
	if got := EncodeToString([]byte{}); got != "" {
		t.Errorf("EncodeToString([]byte{}) = %q, want empty string", got)
	}
}

func TestDecodeString_Empty(t *testing.T) {
	got, err := DecodeString("")
	if err != nil {
		t.Errorf("DecodeString(\"\") returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("DecodeString(\"\") = %v, want empty slice", got)
	}
}

func TestEncodeDecode_Roundtrip(t *testing.T) {
	tests := [][]byte{
		{0x00},
		{0x00, 0x00},
		{1, 2, 3, 4, 5},
		{255, 254, 253},
		[]byte("Hello, World!"),
		[]byte("The quick brown fox jumps over the lazy dog"),
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	for _, tt := range tests {
		encoded := EncodeToString(tt)
		decoded, err := DecodeString(encoded)
		if err != nil {
			t.Errorf("DecodeString(%q) returned error: %v", encoded, err)
			continue
		}
		if len(decoded) != len(tt) {
			t.Errorf("Roundtrip %v: len(decoded) = %d, want %d", tt, len(decoded), len(tt))
			continue
		}
		for i := range tt {
			if decoded[i] != tt[i] {
				t.Errorf("Roundtrip %v: decoded[%d] = %d, want %d", tt, i, decoded[i], tt[i])
			}
		}
	}
}

func TestDecodeString_InvalidChar(t *testing.T) {
	_, err := DecodeString("0OIl") // 0, O, I, l are not in base58 alphabet
	if err != ErrInvalidBase58 {
		t.Errorf("DecodeString with invalid chars: err = %v, want ErrInvalidBase58", err)
	}
}

func TestEncodeToString_LeadingZeros(t *testing.T) {
	// Leading zero bytes should encode to '1' characters
	data := []byte{0, 0, 1, 2, 3}
	encoded := EncodeToString(data)
	if encoded[:2] != "11" {
		t.Errorf("EncodeToString with leading zeros: got %q, want prefix '11'", encoded)
	}

	decoded, err := DecodeString(encoded)
	if err != nil {
		t.Errorf("DecodeString returned error: %v", err)
	}
	if len(decoded) != len(data) {
		t.Errorf("decoded len = %d, want %d", len(decoded), len(data))
	}
	for i := range data {
		if decoded[i] != data[i] {
			t.Errorf("decoded[%d] = %d, want %d", i, decoded[i], data[i])
		}
	}
}

func TestDecodeString_AllOnes(t *testing.T) {
	// All '1's should decode to all zero bytes
	decoded, err := DecodeString("111")
	if err != nil {
		t.Errorf("DecodeString(\"111\") returned error: %v", err)
	}
	for _, b := range decoded {
		if b != 0 {
			t.Errorf("DecodeString(\"111\") should be all zeros, got %d", b)
		}
	}
	if len(decoded) != 3 {
		t.Errorf("DecodeString(\"111\") len = %d, want 3", len(decoded))
	}
}
