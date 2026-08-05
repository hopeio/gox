package base58

import (
	"errors"
	"math/big"
)

const base58Alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte
var ErrInvalidBase58 = errors.New("invalid base58")

// init initializes package state.
func init() {
	for i := 0; i < len(base58Alphabet); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(base58Alphabet); i++ {
		decodeBase58Map[base58Alphabet[i]] = byte(i)
	}
}

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
	base := big.NewInt(58)
	var result []byte

	// Perform radix conversion
	for bigNum.Sign() > 0 {
		remainder := new(big.Int)
		bigNum.DivMod(bigNum, base, remainder)
		result = append(result, base58Alphabet[remainder.Int64()])
	}

	// Append chars for leading zeros (Base58 uses '1')
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	// Reverse the result because digits were computed from low to high
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// DecodeString formats or converts the value.
func DecodeString(s string) ([]byte, error) {
	if s == "" {
		return []byte{}, nil
	}

	// Count leading '1's (matching leading zero bytes in the input)
	leadingOnes := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	// Skip leading '1's and process the rest
	remaining := s[leadingOnes:]
	if remaining == "" {
		// If all '1's, return that many zero bytes
		return make([]byte, leadingOnes), nil
	}

	// Convert a Base58 string to a big integer
	bigNum := new(big.Int)
	base := big.NewInt(58)

	for i := 0; i < len(remaining); i++ {
		char := remaining[i]

		// Find the character index in the Base58 alphabet
		index := int(decodeBase58Map[char])
		if index == 0xFF {
			return nil, ErrInvalidBase58
		}

		// bigNum = bigNum * 58 + index
		bigNum.Mul(bigNum, base)
		bigNum.Add(bigNum, big.NewInt(int64(index)))
	}

	// Convert to a byte slice
	byteArray := bigNum.Bytes()

	// Append leading zero bytes
	if leadingOnes > 0 {
		result := make([]byte, leadingOnes+len(byteArray))
		copy(result[leadingOnes:], byteArray)
		return result, nil
	}

	return byteArray, nil
}
