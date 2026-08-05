/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package math

import (
	"math"
	"time"
)

const magicNumber = 0xf1234fff

// XORing a number with the same value twice yields the original number.

// SecondKey generates a time-based key.
func SecondKey() int64 {
	return time.Now().Unix() ^ magicNumber
}

// ValidateSecondKey returns the absolute difference between the expected key and the provided one.
func ValidateSecondKey(key int64) float64 {
	return math.Abs(float64(key ^ magicNumber - time.Now().Unix()))
}

// GenKey derives a key by XORing with the magic number.
func GenKey(key int64) int64 {
	return key ^ magicNumber
}

// ValidateKey returns the absolute difference between the expected key material and the provided key.
func ValidateKey(key, secretKey int64) float64 {
	return math.Abs(float64(secretKey ^ magicNumber - key))
}
