package base62

import (
	"fmt"
	"math/big"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeToString formats or converts the value.
func EncodeToString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Count leading zeros
	leadingZeros := 0
	for leadingZeros < len(data) && data[leadingZeros] == 0 {
		leadingZeros++
	}

	// Convert bytes to a big integer
	bigNum := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	var result []byte

	for bigNum.Sign() > 0 {
		remainder := new(big.Int)
		bigNum.DivMod(bigNum, base, remainder)
		result = append(result, base62Alphabet[remainder.Int64()])
	}

	// Append characters for leading zeros
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '0') // Leading zeros map to '0' in Base62
	}

	// Reverse the result because digits were computed from low to high
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

var base62DecodeMap [256]byte

// init initializes package state.
func init() {
	for i := range base62DecodeMap {
		base62DecodeMap[i] = 255
	}
	for i, char := range base62Alphabet {
		base62DecodeMap[byte(char)] = byte(i)
	}
}

// DecodeString formats or converts the value.
func DecodeString(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}

	// Count leading zeros
	leadingZeros := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '0' {
			leadingZeros++
		} else {
			break
		}
	}

	// Skip leading zeros and process the rest
	remaining := s[leadingZeros:]
	if remaining == "" {
		// If all zeros
		result := make([]byte, leadingZeros)
		return result, nil
	}

	// Convert a Base62 string to a big integer (alphabet: 0-9a-zA-Z)
	bigNum := new(big.Int)
	base := big.NewInt(62)

	for i := 0; i < len(remaining); i++ {
		char := remaining[i]
		index := base62DecodeMap[char]
		if index == 255 {
			return nil, fmt.Errorf("invalid Base62 character: %c", char)
		}
		bigNum.Mul(bigNum, base)
		bigNum.Add(bigNum, big.NewInt(int64(index)))
	}

	// Convert to a byte slice
	byteArray := bigNum.Bytes()

	// Append leading zeros
	if leadingZeros > 0 {
		result := make([]byte, leadingZeros+len(byteArray))
		copy(result[leadingZeros:], byteArray)
		return result, nil
	}

	return byteArray, nil
}
