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
	"github.com/hopeio/gox/scheduler/light"

	"os"
	"path/filepath"
	"slices"
	"strconv"
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
func Find(path string) (string, error) {
	files, err := FindFiles(path, 8, 1)
	if err != nil {
		return "", err
	}
	return files[0], nil
}

func FindFiles(path string, deep int8, num int) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []string
	filepath1 := filepath.Join(wd, path)
	if _, err = os.Stat(filepath1); !os.IsNotExist(err) {
		files = append(files, filepath1)
		if len(files) == num {
			return files, nil
		}
	}

	subDirFiles(wd, path, "", &files, deep, 0, num)
	supDirFiles(wd+string(os.PathSeparator), path, &files, deep, 0, num)
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
func FindFilesParallel(path string, deep int8, num int) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var file = make(chan string, 1)
	//属于回调而不是通知
	ctx := lightengine.New(context.Background())
	ctx.OnStop(func() { close(file) })
	defer ctx.Cancel()
	// 当前目录下先找
	filepath1 := filepath.Join(wd, path)
	if _, err = os.Stat(filepath1); !os.IsNotExist(err) {
		file <- filepath1
	}

	ctx.AddTask(func() []lightengine.Task {
		subTasks, err := subDirFilesParallel(wd, path, "", file, deep, 0)
		if err != nil {
			log.Error(err)
		}
		return subTasks
	}, func() []lightengine.Task {
		subTasks, err := supDirFilesParallel(wd+PathSeparator, path, file, deep, 0)
		if err != nil {
			log.Error(err)
		}
		return subTasks
	})

	var files []string
	for filepath1 := range file {
		if files = append(files, filepath1); len(files) == num {
			//close(file) 这里无需做关闭操作，会关的
			return files, nil
		}
	}
	return files, nil
}

func subDirFilesParallel(dir, path, exclude string, file chan string, deep, step int8) ([]lightengine.Task, error) {
	if step == deep {
		return nil, nil
	}
	step += 1
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var tasks []lightengine.Task
	for i := range fileInfos {
		if fileInfos[i].IsDir() {
			if exclude != "" && fileInfos[i].Name() == exclude {
				continue
			}
			subDir := filepath.Join(dir, fileInfos[i].Name())
			filepath1 := filepath.Join(subDir, path)
			if _, err = os.Stat(filepath1); !os.IsNotExist(err) {
				file <- filepath1
			}
			tasks = append(tasks, func() []lightengine.Task {
				subTasks, err := subDirFilesParallel(filepath.Join(dir, fileInfos[i].Name()), path, "", file, deep, step)
				if err != nil {
					log.Error(err)
				}
				return subTasks
			})
		}
	}
	return tasks, nil
}

func supDirFilesParallel(dir, path string, file chan string, deep, step int8) ([]lightengine.Task, error) {
	if step == deep {
		return nil, nil
	}
	step += 1
	dir, dirName := filepath.Split(dir[:len(dir)-1])
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	filepath1 := filepath.Join(dir, path)
	if _, err := os.Stat(filepath1); !os.IsNotExist(err) {
		file <- filepath1
	}

	return []lightengine.Task{
		func() []lightengine.Task {
			subTasks, err := subDirFilesParallel(dir, path, dirName, file, deep, 0)
			if err != nil {
				log.Error(err)
			}
			return subTasks
		}, func() []lightengine.Task {
			subTasks, err := supDirFilesParallel(dir, path, file, deep, step)
			if err != nil {
				log.Error(err)
			}
			return subTasks
		},
	}, nil
}

func Mkdir(src string) error {
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		err = os.Mkdir(src, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return err
}

func MkdirAll(src string) error {
	return os.MkdirAll(src, os.ModePerm)
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
	return OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func Open(filepath string) (*os.File, error) {
	return OpenFile(filepath, os.O_RDWR, 0666)
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
