/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"io/fs"
	"iter"
	"os"

	"go.uber.org/multierr"
)

type WalkCallback = func(dir string, entry os.DirEntry) error

// 遍历WalkCallback文件调用callback,包括文件夹,与filepath.WalkDir不同的是回调函数的参数不同,filepath.WalkDir的第一个参数是文件完整路径,WalkFile是文件所在目录的路径
func Walk(dir string, callback WalkCallback) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err1 := callback(dir, entry)
		if err1 != nil {
			if err1 == fs.SkipDir || err1 == fs.SkipAll {
				return nil
			} else {
				err = multierr.Append(err, err1)
			}
		}
		if entry.IsDir() {
			err = multierr.Append(err, Walk(dir+PathSeparator+entry.Name(), callback))
		}
	}

	return err
}

// 遍历根目录中的每个文件，为每个文件调用callback,不包括文件夹,与filepath.WalkDir不同的是回调函数的参数不同,filepath.WalkDir的第一个参数是文件完整路径,WalkFile是文件所在目录的路径
func WalkFile(dir string, callback WalkCallback) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err = multierr.Append(err, WalkFile(dir+PathSeparator+entry.Name(), callback))
		} else {
			err1 := callback(dir, entry)
			if err1 != nil {
				err = multierr.Append(err, err1)
			}
		}
	}

	return err
}

// WalkDir 遍历根目录中的每个文件夹，为文件夹中所有文件和目录的切片(os.ReadDir的返回)调用callback
// callback 需要处理每个文件夹下的所有文件和目录,返回值为需要递归遍历的目录和error
// 几乎每个文件夹下的文件夹都会被循环两次！
func WalkDir(dir string, callback func(dir string, entries []os.DirEntry) ([]os.DirEntry, error)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	dirs, err1 := callback(dir, entries)
	if err1 != nil {
		if err1 == fs.SkipDir || err1 == fs.SkipAll {
			return nil
		}
		err = multierr.Append(err, err1)
	}
	for _, entry := range dirs {
		if entry.IsDir() {
			err = multierr.Append(err, WalkDir(dir+PathSeparator+entry.Name(), callback))
		}
	}
	return err
}

func All(path string) (iter.Seq[os.DirEntry], error) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var errs error
	return func(yield func(os.DirEntry) bool) {
		for _, dir := range dirs {
			if dir.IsDir() {
				it, err := All(path + PathSeparator + dir.Name())
				if err != nil {
					errs = multierr.Append(errs, err)
				}
				for entry := range it {
					if !yield(entry) {
						return
					}
				}
			}
			if !yield(dir) {
				return
			}
		}
	}, errs
}
