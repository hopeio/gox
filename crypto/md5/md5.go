/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package md5

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

// EncodeString formats or converts the value.
func EncodeString(value string) string {
	md5 := md5.Sum([]byte(value))
	return hex.EncodeToString(md5[:])
}

// Encode formats or converts the value.
func Encode(value string) []byte {
	md5 := md5.Sum([]byte(value))
	return md5[:]
}

// ToString returns the string representation.
func ToString(md5 []byte) string {
	return hex.EncodeToString(md5)
}

// EncodeReader formats or converts the value.
func EncodeReader(r io.Reader) ([]byte, error) {
	hash := md5.New()
	_, err := io.Copy(hash, r)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

// EncodeReaderString formats or converts the value.
func EncodeReaderString(r io.Reader) (string, error) {
	hash := md5.New()
	_, err := io.Copy(hash, r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
