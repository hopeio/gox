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
	"path/filepath"

	"go.uber.org/multierr"
)

type RangeCallback = func(dir string, entry os.DirEntry) error
type RangeUntilCallback = func(dir string, entry os.DirEntry) (stop bool, err error)

// 遍历根目录中的每个文件，为每个文件调用callback,包括文件夹,与filepath.WalkDir不同的是回调函数的参数不同,filepath.WalkDir的第一个参数是文件完整路径,RangeFile是文件所在目录的路径
func Range(dir string, callback RangeCallback) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err = multierr.Append(err, RangeFile(dir+PathSeparator+entry.Name(), callback))
		}
		err = multierr.Append(err, callback(dir, entry))
	}

	return err
}

// 指定遍历深度,0为只遍历一层,-1为无限遍历
func RangeDeep(dir string, callback RangeCallback, deep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && deep != 0 {
			err = multierr.Append(err, RangeDeep(dir+PathSeparator+entry.Name(), callback, deep-1))
		}
		err = multierr.Append(err, callback(dir, entry))
	}

	return err
}

// 遍历根目录中的每个文件，为每个文件调用callback,不包括文件夹,与filepath.WalkDir不同的是回调函数的参数不同,filepath.WalkDir的第一个参数是文件完整路径,RangeFile是文件所在目录的路径
func RangeFile(dir string, callback RangeCallback) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err = multierr.Append(err, RangeFile(dir+PathSeparator+entry.Name(), callback))
		} else {
			if rerr := callback(dir, entry); rerr != nil {
				err = multierr.Append(err, rerr)
			}
		}
	}

	return err
}

func RangeFileUntil(dir string, callback RangeUntilCallback) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err = multierr.Append(err, RangeFileUntil(dir+PathSeparator+entry.Name(), callback))
		} else {
			stop, rerr := callback(dir, entry)
			err = multierr.Append(err, rerr)
			if stop {
				return err
			}
		}
	}

	return err
}

// 指定遍历深度,0为只遍历一层,-1为无限遍历
func RangeFileDeep(dir string, callback RangeCallback, deep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() && deep != 0 {
			err = multierr.Append(err, RangeFileDeep(dir+PathSeparator+entry.Name(), callback, deep-1))

		} else {
			err = multierr.Append(err, callback(dir, entry))
		}
	}
	return err
}

// RangeDir 遍历根目录中的每个文件夹，为文件夹中所有文件和目录的切片(os.ReadDir的返回)调用callback
// callback 需要处理每个文件夹下的所有文件和目录,返回值为需要递归遍历的目录和error
// 几乎每个文件夹下的文件夹都会被循环两次！
func RangeDir(dir string, callback func(dir string, entries []os.DirEntry) ([]os.DirEntry, error)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	dirs, err1 := callback(dir, entries)
	err = multierr.Append(err, err1)
	for _, entry := range dirs {
		if entry.IsDir() {
			err = multierr.Append(err, RangeDir(dir+PathSeparator+entry.Name(), callback))
		}
	}
	return err
}

func WalkDirFS(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(fsys, root, fn)
}

func Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

func WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
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
