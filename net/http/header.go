/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Header interface {
	Add(key, value string)
	Set(key, value string)
	Get(key string) string
	Values(key string) []string
	Range(func(key, value string))
}

type IntoHttpHeader interface {
	IntoHttpHeader(header http.Header)
}

// HeaderIntoHttpHeader performs the operation.
func HeaderIntoHttpHeader(header Header, httpHeader http.Header) {
	header.Range(func(key, value string) {
		httpHeader.Add(key, value)
	})
}

// HttpHeaderIntoHeader performs the operation.
func HttpHeaderIntoHeader(httpHeader http.Header, header Header) {
	for k, vv := range httpHeader {
		for _, v := range vv {
			header.Add(k, v)
		}
	}
}

type SliceHeader []string

// NewHeader creates and returns a new instance.
func NewHeader() *SliceHeader {
	h := make(SliceHeader, 0, 6)
	return &h
}

// Add updates or inserts a value.
func (h *SliceHeader) Add(k, v string) {
	*h = append(*h, k, v)

}

// Set updates or inserts a value.
func (h *SliceHeader) Set(k, v string) {
	header := *h
	for i, s := range header {
		if s == k {
			header[i+1] = v
			return
		}
	}
	h.Add(k, v)
}

// Get returns the value.
func (h *SliceHeader) Get(k string) string {
	header := *h
	for i, s := range header {
		if s == k {
			return header[i+1]
		}
	}
	return ""
}

// Values returns the result.
func (h *SliceHeader) Values(k string) []string {
	header := *h
	var values []string
	for i, s := range header {
		if s == k {
			values = append(values, header[i+1])
		}
	}
	return values
}

// Range performs the operation.
func (h *SliceHeader) Range(f func(key, value string)) {
	header := *h
	for i, s := range header {
		if i%2 == 0 {
			f(s, header[i+1])
		}
	}
}

// IntoHttpHeader performs the operation.
func (h SliceHeader) IntoHttpHeader(header http.Header) {
	hlen := len(h)
	for i := 0; i < hlen && i+1 < hlen; i += 2 {
		header.Set(h[i], h[i+1])
	}
}

// ToHttpHeader converts the value.
func (h SliceHeader) ToHttpHeader() http.Header {
	header := make(http.Header)
	h.IntoHttpHeader(header)
	return header
}

// Clone returns the result.
func (h SliceHeader) Clone() SliceHeader {
	newh := make(SliceHeader, len(h))
	copy(newh, h)
	return newh
}

// CopyHttpHeader performs the operation.
func CopyHttpHeader(dst, src http.Header) {
	if src == nil {
		return
	}

	for k, vv := range src {
		for _, v := range vv {
			dst[k] = append(dst[k], v)
		}
	}
}

// ParseDisposition parses the input.
func ParseDisposition(disposition string) (mediatype string, params map[string]string, err error) {
	return mime.ParseMediaType(disposition)
}

// ParseContentRange parses the input.
func ParseContentRange(rangeHeader string) (start int64, end int64, total int64, err error) {
	// Extract the Range value, format "bytes unit-unit/*"
	parts := strings.Split(rangeHeader, " ")
	if len(parts) != 2 || parts[0] != "bytes" {
		err = fmt.Errorf("invalid Content-Range format")
		return
	}
	rangeSpec := parts[1]
	info := strings.Split(rangeSpec, "/")
	bounds := strings.Split(info[0], "-")
	start, err = strconv.ParseInt(bounds[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid range start %w", err)
		return
	}

	if len(bounds) > 1 {
		end, err = strconv.ParseInt(bounds[1], 10, 64)
		if err != nil {
			err = fmt.Errorf("invalid range end %w", err)
			return
		}
	} else {
		// If only a start is given, the end defaults to EOF
		end = -1
	}

	if len(info) == 2 && info[1] != "*" {
		total, err = strconv.ParseInt(info[1], 10, 64)
		if err != nil {
			err = fmt.Errorf("invalid range total %w", err)
			return
		}
	} else {
		total = -1
	}
	return
}

// FormatContentRange formats or converts the value.
// Content-Range 标准格式是 "bytes 0-1/2"（空格分隔）；旧实现输出 "bytes=..."
// 连本包的 ParseContentRange 都解析不了。end=0 是合法值（首字节），按 <0 判缺省。
func FormatContentRange(start, end, total int64) string {
	totalStr := "*"
	if total > 0 {
		totalStr = strconv.FormatInt(total, 10)
	}
	if end < 0 {
		return fmt.Sprintf("bytes %d-/%s", start, totalStr)
	}
	return fmt.Sprintf("bytes %d-%d/%s", start, end, totalStr)
}

// ParseRange parses the input.
func ParseRange(header string) (int64, int64, error) {
	if len(header) < len("bytes=") {
		return 0, 0, fmt.Errorf("invalid Content-Range format")
	}
	header = header[len("bytes="):]
	parts := strings.Split(header, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range header format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	end, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

// FormatRange formats or converts the value.
func FormatRange(start, end int64) string {
	if end <= 0 {
		return fmt.Sprintf("bytes=%d-", start)
	}
	return fmt.Sprintf("bytes=%d-%d", start, end)
}

// ParseContentDisposition parses the input.
func ParseContentDisposition(header string) (string, error) {
	if len(header) <= len("attachment; filename=") {
		return "", fmt.Errorf("invalid Content-Disposition header")
	}
	header = header[len("attachment; filename="):]
	if len(header) >= 2 && header[0] == '"' && header[len(header)-1] == '"' {
		header = header[1 : len(header)-1]
	}
	return url.QueryUnescape(header)
}

// GetContentLength returns the value.
func GetContentLength(header http.Header) int64 {
	length, _ := strconv.ParseInt(header.Get(HeaderContentLength), 10, 64)
	return length
}

// FormatContentDisposition formats or converts the value.
func FormatContentDisposition(filename string) string {
	// Basic example without encoding considerations
	return fmt.Sprintf(`attachment; filename="%s"`, filename)
}

type MapHeader map[string]string

// IntoHttpHeader performs the operation.
func (h MapHeader) IntoHttpHeader(header http.Header) {
	for k, v := range h {
		header.Set(k, v)
	}
}

// ToHttpHeader converts the value.
func (h MapHeader) ToHttpHeader() http.Header {
	header := make(http.Header)
	h.IntoHttpHeader(header)
	return header
}

// Add updates or inserts a value.
func (h MapHeader) Add(k, v string) {
	h[k] = v
}

// Set updates or inserts a value.
func (h MapHeader) Set(k, v string) {
	h[k] = v
}

// Get returns the value.
func (h MapHeader) Get(k string) string {
	return h[k]
}

// Values returns the result.
func (h MapHeader) Values(k string) []string {
	return []string{h[k]}
}

// Range performs the operation.
func (h MapHeader) Range(f func(key, value string)) {
	for k, v := range h {
		f(k, v)
	}
}

type HttpHeader http.Header

// IntoHttpHeader performs the operation.
func (h HttpHeader) IntoHttpHeader(header http.Header) {
	for k, v := range h {
		header.Set(k, v[0])
	}
}

// ToHttpHeader converts the value.
func (h HttpHeader) ToHttpHeader() http.Header {
	return http.Header(h)
}

// Add updates or inserts a value.
func (h HttpHeader) Add(k, v string) {
	http.Header(h).Add(k, v)
}

// Set updates or inserts a value.
func (h HttpHeader) Set(k, v string) {
	http.Header(h).Set(k, v)
}

// Get returns the value.
func (h HttpHeader) Get(k string) string {
	return http.Header(h).Get(k)
}

// Values returns the result.
func (h HttpHeader) Values(k string) []string {
	return http.Header(h).Values(k)
}

// Range performs the operation.
func (h HttpHeader) Range(f func(key, value string)) {
	for k, vv := range h {
		for _, v := range vv {
			f(k, v)
		}
	}
}
