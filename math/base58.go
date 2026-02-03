package math

import (
	"errors"
)

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
func init() {

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte
var ErrInvalidBase58 = errors.New("invalid base58")

// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte
var ErrInvalidBase32 = errors.New("invalid base32")

const encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"

var decodeBase32Map [256]byte

const encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte

// FormatUintBase32 uses the z-base-32 character set but encodes and decodes similar
// to base58, allowing it to create an even smaller result string.
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func FormatUintBase32(f uint64) string {

	if f < 32 {
		return string(encodeBase32Map[f])
	}

	b := make([]byte, 0, 12)
	for f >= 32 {
		b = append(b, encodeBase32Map[f%32])
		f /= 32
	}
	b = append(b, encodeBase32Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase32Uint parses a base32 []byte into a uint64
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func ParseBase32Uint(b []byte) (uint64, error) {
	var id uint64
	for i := range b {
		if decodeBase32Map[b[i]] == 0xFF {
			return 0, ErrInvalidBase32
		}
		id = id*32 + uint64(decodeBase32Map[b[i]])
	}

	return id, nil
}

// FormatUintBase58 returns a base58 string of the uint64
func FormatUintBase58(f uint64) string {

	if f < 58 {
		return string(encodeBase58Map[f])
	}

	b := make([]byte, 0, 11)
	for f >= 58 {
		b = append(b, encodeBase58Map[f%58])
		f /= 58
	}
	b = append(b, encodeBase58Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase58Uint parses a base58 []byte into a uint64
func ParseBase58Uint(b []byte) (uint64, error) {
	var id uint64
	for i := range b {
		if decodeBase58Map[b[i]] == 0xFF {
			return 0, ErrInvalidBase58
		}
		id = id*58 + uint64(decodeBase58Map[b[i]])
	}

	return id, nil
}
