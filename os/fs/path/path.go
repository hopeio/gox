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

// FileRewrite 删除文件名中的非法字符（/ \ * |），并把 < > ? : 替换为全角对应符号。
// 旧实现丢弃所有普通字符、replace 表只填了一个元素（idx≥1 时越界 panic）。
func FileRewrite(filename string) string {
	empty := []rune{'/', '\\', '*', '|'}
	origin := []rune{'<', '>', '?', ':'}
	replace := []rune{'《', '》', '？', '：'}
	result := make([]rune, 0, len(filename))
	for _, char := range filename {
		if slices.Contains(empty, char) {
			continue
		}
		if idx := slices.Index(origin, char); idx >= 0 {
			result = append(result, replace[idx])
			continue
		}
		result = append(result, char)
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
// 旧实现两个空值分支互相拿错参数，且用 path[len(dir)-1-len(file)]（常为负索引）取分隔符。
func Cleanse(path string) string { // will be used when save the dir or the part
	dir, file := filepath.Split(path)
	if dir == "" {
		return FileCleanse(file)
	}
	if file == "" {
		return DirCleanse(dir)
	}
	// filepath.Split 的 dir 带尾分隔符：去掉后分别清洗，再用原分隔符拼回
	sep := dir[len(dir)-1]
	return DirCleanse(dir[:len(dir)-1]) + string(sep) + FileCleanse(file)
}

// FileNoExt returns the result.
func FileNoExt(filepath string) string {
	base := sdpath.Base(filepath)
	return base[:len(base)-len(sdpath.Ext(base))]
}
