package multibase

import (
	"math/big"
	"testing"
)

// TestEncodeBase62 tests the EncodeBase62 function with various inputs
func TestEncodeBase62(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string // Test case name
		input    []byte // Input byte array to encode
		expected string // Expected Base62 encoded string
	}{
		{
			name:     "Empty input", // Empty input should return empty string
			input:    []byte{},
			expected: "",
		},
		{
			name:     "Single zero byte", // Single zero byte should return "0"
			input:    []byte{0},
			expected: "0",
		},
		{
			name:     "Multiple zero bytes", // Multiple zero bytes should return multiple "0"s
			input:    []byte{0, 0, 0},
			expected: "000",
		},
		{
			name:     "Single non-zero byte", // Single non-zero byte (1) should return "1"
			input:    []byte{1},
			expected: "1",
		},
		{
			name:     "Two bytes - small value", // Two bytes representing a small number
			input:    []byte{0, 1},              // Should have leading zero
			expected: "01",
		},
		{
			name:     "Two bytes - larger value", // Two bytes representing a larger number
			input:    []byte{1, 0},               // Equals 256 in decimal
			expected: "4c",                       // 256 in base62 is "4c"
		},
		{
			name:     "Value equals 62", // Value that equals base (62) should be "10"
			input:    []byte{62},        // 62 in decimal
			expected: "10",              // 62 in base62 is "10"
		},
		{
			name:     "Value equals 61", // Value just below base (61) should be single char
			input:    []byte{61},        // 61 in decimal
			expected: "Z",               // 61 maps to 'z' in base62 alphabet
		},
		{
			name:     "Large value requiring multiple digits", // Large number needing multiple base62 digits
			input:    []byte{255, 255},                        // 0xFFFF = 65535 in decimal
			expected: "h31",                                   // 65535 in base62
		},
		{
			name:     "Leading zeros with non-zero content", // Array with leading zeros and actual data
			input:    []byte{0, 0, 255},                     // Leading zeros followed by 255
			expected: "0047",                                // Leading zeros preserved, 255 encoded as "4V"
		},
		{
			name:     "Maximum single byte value", // Maximum value for single byte (255)
			input:    []byte{255},                 // 255 in decimal
			expected: "47",                        // 255 in base62 is "47"
		},
		{
			name:     "Three bytes forming large number", // Three bytes forming a large number
			input:    []byte{1, 0, 0},                    // 0x010000 = 65536 in decimal
			expected: "h32",                              // 65536 in base62
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			result := EncodeBase62(tt.input)

			// Compare result with expected value
			if result != tt.expected {
				t.Errorf("EncodeBase62(%v) = %s; expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEncodeBase62WithKnownValues tests with known conversions to verify correctness
func TestEncodeBase62WithKnownValues(t *testing.T) {
	// Test some known conversions manually calculated or verified
	knownConversions := map[string][]byte{
		"0":   {0},
		"1":   {1},
		"a":   {10},         // 10 in base62 alphabet is 'A'
		"z":   {35},         // 35 in base62 alphabet is 'Z'
		"A":   {36},         // 36 in base62 alphabet is 'a'
		"Z":   {61},         // 61 in base62 alphabet is 'z'
		"10":  {62},         // 62 = 1*62 + 0
		"11":  {63},         // 63 = 1*62 + 1
		"100": {0x0f, 0x04}, // 3844 = 1*62^2 + 0*62 + 0
	}

	for expected, input := range knownConversions {
		result := EncodeBase62(input)
		if result != expected {
			t.Errorf("EncodeBase62(%v) = %s; expected %s", input, result, expected)
		}
	}
}

// TestEncodeBase62Consistency tests that encoding and decoding are consistent
// This assumes there's a corresponding DecodeBase62 function
func TestEncodeBase62Consistency(t *testing.T) {
	// Test consistency with math/big operations
	testCases := [][]byte{
		{0},
		{1},
		{62},
		{255},
		{0, 1},
		{1, 0},
		{255, 255},
		{1, 2, 3, 4, 5},
		{0, 0, 255},
		{0, 1, 0, 1},
	}

	for _, testCase := range testCases {
		encoded := EncodeBase62(testCase)

		// Verify that the encoded string can be converted back to the same number
		// using math/big operations
		originalBig := new(big.Int).SetBytes(testCase)

		// Manually decode the base62 string to verify
		decodedBig := new(big.Int)
		base := big.NewInt(62)

		for _, r := range encoded {
			decodedBig.Mul(decodedBig, base) // decodedBig *= 62

			// Find the index of this character in base62Alphabet
			var digit int64
			switch {
			case r >= '0' && r <= '9':
				digit = int64(r - '0')
			case r >= 'A' && r <= 'Z':
				digit = int64(r - 'A' + 10)
			case r >= 'a' && r <= 'z':
				digit = int64(r - 'a' + 36)
			default:
				t.Fatalf("Invalid base62 character: %c", r)
			}

			decodedBig.Add(decodedBig, big.NewInt(digit)) // decodedBig += digit
		}

		if decodedBig.Cmp(originalBig) != 0 {
			t.Errorf("Encode/decode mismatch for %v: encoded to '%s', but re-decoded to %s instead of original %s",
				testCase, encoded, decodedBig.String(), originalBig.String())
		}
	}
}
