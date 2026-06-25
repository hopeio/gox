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
	"slices"
	"strconv"
	"sync"
)

const (
	ModeDir  = 0755
	ModeFile = 0644
)

type Dir string

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

// path和filepath两个包，filepath文件专用
func Find(src, dst string) (string, error) {
	files, err := FindFiles(src, dst, 8, 1)
	if err != nil {
		return "", err
	}
	return files[0], nil
}

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
		return nil, errors.New("找不到文件")
	}
	return files, nil
}

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

// path和filepath两个包，filepath文件专用
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

	// 当前目录下先找
	filepath1 := filepath.Join(src, dst)
	if _, err := os.Stat(filepath1); !os.IsNotExist(err) {
		file <- filepath1
	}

	// 并行搜索子目录和父目录
	wg.Add(2)
	go func() {
		defer wg.Done()
		subDirFilesParallel(ctx, src, dst, "", file, deep, 0)
	}()
	go func() {
		defer wg.Done()
		supDirFilesParallel(ctx, src+PathSeparator, dst, file, deep, 0)
	}()

	// 所有搜索完成后关闭 channel
	go func() {
		wg.Wait()
		close(file)
	}()

	var files []string
	for fp := range file {
		files = append(files, fp)
		if len(files) == num {
			cancel() // 停止剩余搜索
			break
		}
	}
	if len(files) == 0 {
		return nil, errors.New("找不到文件")
	}
	return files, nil
}

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

func MkdirAll(src string) error {
	return os.MkdirAll(src, ModeDir)
}

func IsExist(src string) bool {
	_, err := os.Stat(src)
	return !os.IsNotExist(err)
}

func IsNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func IsPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

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

func Create(filepath string) (*os.File, error) {
	return OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ModeFile)
}

func Open(filepath string) (*os.File, error) {
	return OpenFile(filepath, os.O_RDWR, ModeFile)
}

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

// LastFile 当前目录最后一个创建的文件
func LastFile(dir string) (os.FileInfo, map[string]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if len(entries) == 0 {
		return nil, nil, err
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		filei, _ := a.Info()
		filej, _ := b.Info()
		return filei.ModTime().Compare(filej.ModTime())
	})
	lastFile, err := entries[0].Info()
	if err != nil {
		return nil, nil, err
	}
	m := make(map[string]os.FileInfo)
	for _, entity := range entries {
		m[entity.Name()], _ = entity.Info()
	}
	return lastFile, m, nil
}

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
		*f = FileSize(size * 1024)
	default:
		unitLen = 1
		size, err := strconv.Atoi(string(text[:len(text)-unitLen]))
		if err != nil {
			return err
		}
		*f = FileSize(size * 8)
	}
	return nil
}
