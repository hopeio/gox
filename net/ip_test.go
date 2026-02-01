package net

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUint64ToIP tests the Uint64ToIP function with various input cases
func TestUint64ToIP(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected net.IP
	}{
		{
			name:     "input is zero should return nil",
			input:    0,
			expected: nil,
		},
		{
			name:     "small positive value (IPv4)",
			input:    1,
			expected: net.IP{0, 0, 0, 1}, // Little endian interpretation
		},
		{
			name:     "max uint32 value (IPv4 boundary)",
			input:    4294967295, // math.MaxUint32
			expected: net.IP{255, 255, 255, 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Uint32ToIPv4(tt.input)

			// Compare the results
			if !assert.Equal(t, result, tt.expected) {
				t.Errorf("Uint64ToIP(%d) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIPv4(t *testing.T) {
	t.Log([]byte(net.IPv4(1, 1, 1, 1)))
}
