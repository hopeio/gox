/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	stringsx "github.com/hopeio/gox/strings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sourceDir string

// init initializes package state.
func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	sourceDir = regexp.MustCompile(`log.utils\.go`).ReplaceAllString(file, "")
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, sourceDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

// trimLineBreak returns the result.
func trimLineBreak(path string) string {
	return path[:len(path)-1]
}

// TrimLineBreak returns the result.
func TrimLineBreak(path string) string {
	if b := path[len(path)-1]; b == '\n' {
		return path[:len(path)-1]
	}
	return path
}

// sprintln returns the result.
func sprintln(a ...any) string {
	return trimLineBreak(fmt.Sprintln(a...))
}

// getMessage returns the result.
func getMessage(fmtArgs []interface{}) string {
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}

// RawJson returns the result.
func RawJson(key string, data []byte) zapcore.Field {
	return zap.Reflect(key, rawJson{Data: data})
}

type rawJson struct {
	Data []byte
}

// MarshalJSON returns m as the JSON encoding of m.
func (m rawJson) MarshalJSON() ([]byte, error) {
	if m.Data == nil {
		return []byte("null"), nil
	}
	return m.Data, nil
}

// String returns the string representation.
func (m rawJson) String() string {
	return stringsx.FromBytes(m.Data)
}
