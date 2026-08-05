/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package encoding

import (
	"bytes"
	"io"
	"strings"

	stringsx "github.com/hopeio/gox/strings"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// DetermineEncoding ...
func DetermineEncoding(content []byte, contentType string) (e encoding.Encoding, name string, certain bool) {
	return charset.DetermineEncoding(content, contentType)
}

// GBKToUTF8 ...
func GBKToUTF8(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return stringsx.FromBytes(b), nil
}

// GBKBytesToUTF8 ...
func GBKBytesToUTF8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	return io.ReadAll(reader)
}

// UTF-8 转 GBK

// UTF8ToGBK ...
func UTF8ToGBK(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return stringsx.FromBytes(b), nil
}

// UTF8BytesToGBK ...
func UTF8BytesToGBK(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	return io.ReadAll(reader)
}
