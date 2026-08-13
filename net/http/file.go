/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Fs = http.FileSystem

type File struct {
	http.File
}

type FileInfo struct {
	name    string
	modTime time.Time
	size    int64
	mode    fs.FileMode
	Body    io.ReadCloser
}

// Name returns the result.
func (f *FileInfo) Name() string {
	return f.name
}

// Size returns the number of elements.
func (f *FileInfo) Size() int64 {
	return f.size
}

// Mode returns the result.
func (f *FileInfo) Mode() fs.FileMode {
	return f.mode
}

// ModTime returns the result.
func (f *FileInfo) ModTime() time.Time {
	return f.modTime
}

// IsDir reports whether the condition holds.
func (f *FileInfo) IsDir() bool {
	return false
}

// Sys returns the result.
func (f *FileInfo) Sys() any {
	return nil
}

// GetFileExt returns the value.
func GetFileExt(file *multipart.FileHeader) (string, error) {
	var ext string
	var index = strings.LastIndex(file.Filename, ".")
	if index == -1 {
		return "", nil
	} else {
		ext = file.Filename[index:]
	}
	if len(ext) == 1 {
		return "", errors.New("invalid extension")
	}
	return ext, nil
}

// CheckFileSize reports whether the condition holds.
func CheckFileSize(f multipart.File, uploadMaxSize int) bool {
	size := GetFileSize(f)
	if size == 0 {
		return false
	}

	return size <= uploadMaxSize
}

// GetFileSize returns the value.
func GetFileSize(f multipart.File) int {
	content, err := io.ReadAll(f)
	if err != nil {
		return 0
	}
	return len(content)
}

// FetchFile performs the operation.
func FetchFile(url string, options ...func(r *http.Request)) (*FileInfo, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for _, option := range options {
		option(req)
	}
	return FetchFileByRequest(req)
}

// FetchFileByRequest performs the operation.
func FetchFileByRequest(r *http.Request) (*FileInfo, error) {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New(resp.Status)
	}

	var file FileInfo
	file.Body = resp.Body
	file.name = path.Base(resp.Request.URL.Path)
	file.modTime, _ = time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	file.size, _ = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	return &file, nil
}
