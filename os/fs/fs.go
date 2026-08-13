/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/hopeio/gox/log"

	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	ModeDir  = 0755
	ModeFile = 0644
)

type Dir string

// Open performs the operation.
func (d Dir) Open(name string) (*os.File, error) {
	dir := string(d)
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(filepath.Clean(string(os.PathSeparator)+name)))
	f, err := os.Open(fullName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Find performs the operation.
func Find(src, dst string) (string, error) {
	files, err := FindFiles(src, dst, 8, 1)
	if err != nil {
		return "", err
	}
	return files[0], nil
}

// FindFiles performs the operation.
func FindFiles(src, dst string, deep int8, num int) ([]string, error) {
	if src == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		src = wd
	}

	var files []string
	filepath1 := filepath.Join(src, dst)
	if _, err := os.Stat(filepath1); !os.IsNotExist(err) {
		files = append(files, filepath1)
		if len(files) == num {
			return files, nil
		}
	}

	subDirFiles(src, dst, "", &files, deep, 0, num)
	supDirFiles(src+string(os.PathSeparator), dst, &files, deep, 0, num)
	if len(files) == 0 {
		return nil, errors.New("file not found")
	}
	return files, nil
}

// subDirFiles performs the operation.
func subDirFiles(dir, path, exclude string, files *[]string, deep, step int8, num int) {
	step += 1
	if step-1 == deep {
		return
	}
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		log.Error(err)
	}
	for i := range fileInfos {
		if fileInfos[i].IsDir() {
			if exclude != "" && fileInfos[i].Name() == exclude {
				continue
			}
			filepath1 := filepath.Join(dir, fileInfos[i].Name(), path)
			if _, err = os.Stat(filepath1); !os.IsNotExist(err) {
				*files = append(*files, filepath1)
				if len(*files) == num {
					return
				}
			}
			subDirFiles(filepath.Join(dir, fileInfos[i].Name()), path, "", files, deep, step, num)
		}
	}
}

// supDirFiles performs the operation.
func supDirFiles(dir, path string, files *[]string, deep, step int8, num int) {
	step += 1
	if step-1 == deep {
		return
	}
	dir, dirName := filepath.Split(dir[:len(dir)-1])
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}
	filepath1 := filepath.Join(dir, path)
	if _, err := os.Stat(filepath1); !os.IsNotExist(err) {
		*files = append(*files, filepath1)
		if len(*files) == num {
			return
		}
	}
	subDirFiles(dir, path, dirName, files, deep, 0, num)
	supDirFiles(dir, path, files, deep, step, num)
}

// FindFilesParallel performs the operation.
func FindFilesParallel(src, dst string, deep int8, num int) ([]string, error) {
	if src == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		src = wd
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file := make(chan string, num+1)
	var wg sync.WaitGroup

	// Search the current directory first
	filepath1 := filepath.Join(src, dst)
	if _, err := os.Stat(filepath1); !os.IsNotExist(err) {
		file <- filepath1
	}

	// Search subdirs and parents in parallel
	wg.Add(2)
	go func() {
		defer wg.Done()
		subDirFilesParallel(ctx, src, dst, "", file, deep, 0)
	}()
	go func() {
		defer wg.Done()
		supDirFilesParallel(ctx, src+PathSeparator, dst, file, deep, 0)
	}()

	// Close the channel after all searches finish
	go func() {
		wg.Wait()
		close(file)
	}()

	var files []string
	for fp := range file {
		files = append(files, fp)
		if len(files) == num {
			cancel() // Cancel remaining searches
			break
		}
	}
	if len(files) == 0 {
		return nil, errors.New("file not found")
	}
	return files, nil
}

// subDirFilesParallel performs the operation.
func subDirFilesParallel(ctx context.Context, dir, path, exclude string, file chan<- string, deep, step int8) {
	if step >= deep || ctx.Err() != nil {
		return
	}
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		log.Error(err)
		return
	}
	var wg sync.WaitGroup
	for i := range fileInfos {
		if !fileInfos[i].IsDir() {
			continue
		}
		if exclude != "" && fileInfos[i].Name() == exclude {
			continue
		}
		subDir := filepath.Join(dir, fileInfos[i].Name())
		fp := filepath.Join(subDir, path)
		if _, err := os.Stat(fp); !os.IsNotExist(err) {
			select {
			case file <- fp:
			case <-ctx.Done():
				wg.Wait()
				return
			}
		}
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			subDirFilesParallel(ctx, d, path, "", file, deep, step+1)
		}(subDir)
	}
	wg.Wait()
}

// supDirFilesParallel performs the operation.
func supDirFilesParallel(ctx context.Context, dir, path string, file chan<- string, deep, step int8) {
	if step >= deep || ctx.Err() != nil {
		return
	}
	dir, dirName := filepath.Split(dir[:len(dir)-1])
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}
	fp := filepath.Join(dir, path)
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		select {
		case file <- fp:
		case <-ctx.Done():
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		subDirFilesParallel(ctx, dir, path, dirName, file, deep, 0)
	}()
	go func() {
		defer wg.Done()
		supDirFilesParallel(ctx, dir, path, file, deep, step+1)
	}()
	wg.Wait()
}

// Mkdir performs the operation.
func Mkdir(src string) error {
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		err = os.Mkdir(src, ModeDir)
		if err != nil {
			return err
		}
	}
	return err
}

// MkdirAll performs the operation.
func MkdirAll(src string) error {
	return os.MkdirAll(src, ModeDir)
}

// IsExist reports whether the condition holds.
func IsExist(src string) bool {
	_, err := os.Stat(src)
	return !os.IsNotExist(err)
}

// IsNotExist reports whether the condition holds.
func IsNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

// IsPermission reports whether the condition holds.
func IsPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// MustOpen returns the value.
func MustOpen(filePath string) (*os.File, error) {
	perm := IsPermission(filePath)
	if !perm {
		return os.Create(filePath)
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("os.Getwd err: %v", err)
	}

	src := dir + PathSeparator + filePath

	return Create(src)
}

// Create performs the operation.
func Create(filepath string) (*os.File, error) {
	return OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ModeFile)
}

// Open performs the operation.
func Open(filepath string) (*os.File, error) {
	return OpenFile(filepath, os.O_RDWR, ModeFile)
}

// OpenFile creates and returns a new instance.
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		dir := filepath.Clean(filepath.Dir(path))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, flag, perm)
}

// LastFile performs the operation.
func LastFile(dir string) (os.FileInfo, map[string]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if len(entries) == 0 {
		return nil, nil, err
	}
	// O(n) 找最新修改的文件即可，无需排序；旧实现升序取 [0] 返回的是最旧的，
	// 且忽略 Info() 错误会在文件被并发删除时 nil panic
	m := make(map[string]os.FileInfo, len(entries))
	var last os.FileInfo
	for _, entry := range entries {
		info, ierr := entry.Info()
		if ierr != nil {
			continue // 条目可能在 ReadDir 后被删除
		}
		m[entry.Name()] = info
		if last == nil || info.ModTime().After(last.ModTime()) {
			last = info
		}
	}
	if last == nil {
		return nil, nil, fmt.Errorf("no readable entries in %s", dir)
	}
	return last, m, nil
}

// Move performs the operation.
func Move(src, dst string) error {
	dir := filepath.Clean(filepath.Dir(dst))
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return os.Rename(src, dst)
}

type FileSize int64

// MarshalText
func (f FileSize) MarshalText() ([]byte, error) {
	buffer := bytes.NewBufferString("")
	if f/FileSize(1024*1024*1024*8) > 0 {
		buffer.WriteString(fmt.Sprintf("%.2f", float64(f)/float64(1024*1024*1024*8)))
		buffer.WriteString("GB")
	} else if f/FileSize(1024*1024*8) > 0 {
		buffer.WriteString(fmt.Sprintf("%.2f", float64(f)/float64(1024*1024*8)))
		buffer.WriteString("MB")
	} else if f/FileSize(1024*8) > 0 {
		buffer.WriteString(fmt.Sprintf("%.2f", float64(f)/float64(1024*8)))
		buffer.WriteString("KB")
	} else {
		buffer.WriteString(fmt.Sprintf("%d", f/8))
		buffer.WriteString("B")
	}
	return buffer.Bytes(), nil
}

// UnMarshalText
func (f *FileSize) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	unitLen := 1
	unit := text[len(text)-1]
	if unit == 'B' {
		unit = text[len(text)-2]
	}
	switch unit {
	case 'G', 'g':
		unitLen = 2
		size, err := strconv.Atoi(string(text[:len(text)-unitLen]))
		if err != nil {
			return err
		}
		*f = FileSize(size * 1024 * 1024 * 1024 * 8)
	case 'M', 'm':
		unitLen = 2
		size, err := strconv.Atoi(string(text[:len(text)-unitLen]))
		if err != nil {
			return err
		}
		*f = FileSize(size * 1024 * 1024 * 8)
	case 'K', 'k':
		unitLen = 2
		size, err := strconv.Atoi(string(text[:len(text)-unitLen]))
		if err != nil {
			return err
		}
		// 内部单位为比特，与 G/M 分支及 MarshalText 保持一致（曾漏乘 8 导致往返不一致）
		*f = FileSize(size * 1024 * 8)
	default:
		if unit >= '0' && unit <= '9' {
			// 纯数字（含 "100B"）不能按单位字符砍位
			unitLen = 0
			if text[len(text)-1] == 'B' || text[len(text)-1] == 'b' {
				unitLen = 1
			}
		}
		size, err := strconv.Atoi(string(text[:len(text)-unitLen]))
		if err != nil {
			return err
		}
		*f = FileSize(size * 8)
	}
	return nil
}
