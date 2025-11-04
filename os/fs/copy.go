/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"io"
	"os"
	"syscall"

	"github.com/hopeio/gox/crypto/md5"
)

type mode int

const (
	Cover mode = iota
	SameNameSkip
	SameNameAndMd5Skip
	// TODO
	sameNameRename
)

func (c mode) handle(dst string, src io.Reader) (skip bool, err error) {
	switch c {
	case Cover:
		return false, nil
	case SameNameSkip:
		return IsExist(dst), nil
	case SameNameAndMd5Skip:
		if IsExist(dst) {
			md51, err := Md5(dst)
			if err != nil {
				return false, err
			}
			md52, err := md5.EncodeReaderString(src)
			if err != nil {
				return false, err
			}

			if md51 == md52 {
				return true, nil
			}
		}
	}
	return false, nil
}

func Copy(dst, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	return CreateFromReader(dst, r)
}

func CopyByMode(dst, src string, c mode) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()
	skip, err := c.handle(dst, r)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	return CreateFromReader(dst, r)
}

const DownloadKey = ".downloading"

func CreateFromReader(filepath string, reader io.Reader) error {
	f, err := Create(filepath)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, reader); err != nil {
		f.Close()
		os.Remove(filepath)
		return err
	}

	if err = f.Close(); err != nil {
		os.Remove(filepath)
		return err
	}
	return nil
}

func CreateFromReaderByMode(filepath string, reader io.Reader, c mode) error {
	skip, err := c.handle(filepath, reader)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	return CreateFromReader(filepath, reader)
}

func Download(filepath string, reader io.Reader) error {
	tmpFilepath := filepath + DownloadKey
	err := CreateFromReader(tmpFilepath, reader)
	if err != nil {
		return err
	}
	return os.Rename(tmpFilepath, filepath)
}

func DownloadByMode(filepath string, reader io.Reader, c mode) error {
	skip, err := c.handle(filepath, reader)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	return Download(filepath, reader)
}

// CopyDirByMode 递归复制目录
func CopyDirByMode(dst, src string, c mode) error {
	if src[len(src)-1] == os.PathSeparator {
		src = src[:len(src)-1]
	}
	if dst[len(dst)-1] == os.PathSeparator {
		dst = dst[:len(dst)-1]
	}
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	for _, entry := range entries {
		entityName := entry.Name()
		if entry.IsDir() {
			err = CopyDirByMode(dst+PathSeparator+entityName, src+PathSeparator+entityName, c)
			if err != nil {
				return err
			}
		} else {
			err = CopyByMode(dst+PathSeparator+entityName, src+PathSeparator+entityName, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyDir 递归复制目录
func CopyDir(dst, src string) error {
	return CopyDirByMode(dst, src, Cover)
}

func MoveDirByMode(src, dst string, c mode) error {
	if src[len(src)-1] == os.PathSeparator {
		src = src[:len(src)-1]
	}
	if dst[len(dst)-1] == os.PathSeparator {
		dst = dst[:len(dst)-1]
	}
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return syscall.Rmdir(src)
	}
	for _, entry := range entries {
		entityName := entry.Name()
		if entry.IsDir() {
			err = MoveDirByMode(src+PathSeparator+entityName, dst+PathSeparator+entityName, c)
			if err != nil {
				return err
			}
		} else {
			skip, err := c.handle(dst+PathSeparator+entityName, nil)
			if err != nil {
				return err
			}
			if skip {
				continue
			}
			err = os.Rename(src+PathSeparator+entityName, dst+PathSeparator+entityName)
			if err != nil {
				return err
			}
		}
	}
	entries, err = os.ReadDir(src)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return syscall.Rmdir(src)
	}
	return nil
}
