/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package path

import (
	sdpath "path"
	"path/filepath"
	"slices"
	"strings"

	stringsx "github.com/hopeio/gox/strings"
)

// FileRewrite returns the result.
func FileRewrite(filename string) string {
	var result []rune
	var empty = []rune{'/', '\\', '*', '|'}
	origin := []rune{'<', '>', '?', ':'}
	var replace []rune
	for _, char := range "《》？：" {
		replace = append(result, char)
	}

	for _, char := range filename {
		if slices.Contains(empty, char) {
			continue
		}
		if idx := slices.Index(origin, char); idx >= 0 {
			result = append(result, replace[idx])
		}
	}
	return string(result)
}

// FileCleanse returns the result.
func FileCleanse(filename string) string {

	filename = strings.Trim(filename, ".-+")
	// windows
	//filename = stringsx.RemoveRunes(filename, '/', '\\', ':', '*', '?', '"', '<', '>', '|')
	// linux
	//filename = stringsx.RemoveRunes(filename, '\'', '*','?', '@', '#', '$', '&', '(', ')', '|', ';',  '/', '%', '^', ' ', '\t', '\n')

	filename = stringsx.RemoveRunes(filename, '/', '\\', ':', '*', '?', '"', '<', '>', '|', ';', '/', '%', '^', ' ', '\t', '\n', '$', '&')
	// CJK punctuation
	//filename = stringsx.RemoveRunes(filename, '：', '，', '。', '！', '？', '、', '“', '”', '、')
	return filename
}

// DirCleanse returns the result.
func DirCleanse(dir string) string { // will be used when save the dir or the part
	// remove special symbol
	// : allowed on unix; required handling on windows
	// windows path
	if len(dir) > 2 && dir[1] == ':' && ((dir[0] >= 'A' && dir[0] <= 'Z') || (dir[0] >= 'a' && dir[0] <= 'z')) && (dir[2] == '/' || dir[2] == '\\') {
		return dir[:3] + stringsx.RemoveRunes(dir[3:], ':', '*', '?', '"', '<', '>', '|', ',', ' ', '\t', '\n')
	}
	return stringsx.RemoveRunes(dir, ':', '*', '?', '"', '<', '>', '|', ',', ' ', '\t', '\n')
}

// Cleanse returns the result.
func Cleanse(path string) string { // will be used when save the dir or the part
	dir, file := filepath.Split(path)
	if dir == "" {
		return DirCleanse(dir)
	}
	if file == "" {
		return FileCleanse(file)
	}
	// remove special symbol
	return DirCleanse(dir) + string(path[len(dir)-1-len(file)]) + FileCleanse(file)
}

// FileNoExt returns the result.
func FileNoExt(filepath string) string {
	base := sdpath.Base(filepath)
	return base[:len(base)-len(sdpath.Ext(base))]
}
