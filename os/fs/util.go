/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	stdpath "path"
	"strings"

	md52 "github.com/hopeio/gox/crypto/md5"
	"github.com/hopeio/gox/log"
	"github.com/hopeio/gox/slices"
)

func Exist(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func NotExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return os.IsNotExist(err)
}

func Md5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		file.Close()
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func Md5Equal(path1, path2 string) (bool, error) {
	md51, err := Md5(path1)
	if err != nil {
		return false, err
	}
	md52, err := Md5(path2)
	if err != nil {
		return false, err
	}
	return md51 == md52, nil
}

func GetMd5Name(name string) string {
	ext := stdpath.Ext(name)
	fileName := strings.TrimSuffix(name, ext)
	fileName = md52.EncodeString(fileName)
	return fileName + ext
}

type duplicateFile struct {
	path string
	md5  string
}

// 去除目录中重复的文件,默认保留参数靠前目录中的文件
func DirsDeDuplicate(dirs ...string) error {
	return DirsDuplicateHandle(func(path1, path2 string) error {
		log.Debugf("exists: %s,remove:%s", path1, path2)
		return os.Remove(path2)
	}, dirs...)
}

func DirsDuplicateHandle(callback func(path1, path2 string) error, dirs ...string) error {
	fileSizeMap := make(map[int64][]*duplicateFile)
	for _, tmpDir := range dirs {
		err := RangeFile(tmpDir, func(dir string, entry os.DirEntry) error {
			info, _ := entry.Info()
			path := dir + PathSeparator + entry.Name()
			duplicateFiles, ok := fileSizeMap[info.Size()]
			var entryMd5 string
			if ok {
				var err error
				entryMd5, err = Md5(path)
				if err != nil {
					return err
				}
				for _, file := range duplicateFiles {
					if file.md5 == "" {
						file.md5, err = Md5(file.path)
						if err != nil {
							return err
						}
					}
					if file.md5 == entryMd5 {
						return callback(file.path, path)
					}
				}
			}
			fileSizeMap[info.Size()] = append(fileSizeMap[info.Size()], &duplicateFile{path: path, md5: entryMd5})
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DirsRangeDuplicateHandle
// rangeCallback 返回值为错误和是否继续
func DirsRangeDuplicateHandle(rangeCallback func(dir string, entry os.DirEntry) (error, bool), duplicateCallback func(path1, path2 string) error, dirs ...string) error {
	fileSizeMap := make(map[int64][]*duplicateFile)
	for _, tmpDir := range dirs {
		err := RangeFile(tmpDir, func(dir string, entry os.DirEntry) error {
			if err, goon := rangeCallback(dir, entry); !goon {
				return err
			}

			info, _ := entry.Info()
			path := dir + PathSeparator + entry.Name()
			duplicateFiles, ok := fileSizeMap[info.Size()]
			var entryMd5 string
			if ok {
				var err error
				entryMd5, err = Md5(path)
				if err != nil {
					return err
				}
				for _, file := range duplicateFiles {
					if file.md5 == "" {
						file.md5, err = Md5(file.path)
						if err != nil {
							return err
						}
					}
					if file.md5 == entryMd5 {
						return duplicateCallback(file.path, path)
					}
				}
			}
			fileSizeMap[info.Size()] = append(fileSizeMap[info.Size()], &duplicateFile{path: path, md5: entryMd5})
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func TwoDirDuplicateHandle(dir1, dir2 string, callback func(path1, path2 string) error) error {
	fileSizeMap := make(map[int64][]*duplicateFile)
	err := RangeFile(dir1, func(dir string, entry os.DirEntry) error {
		info, _ := entry.Info()
		fileSizeMap[info.Size()] = append(fileSizeMap[info.Size()], &duplicateFile{path: dir + PathSeparator + entry.Name()})
		return nil
	})

	if err != nil {
		return err
	}

	return RangeFile(dir2, func(dir string, entry os.DirEntry) error {
		info, _ := entry.Info()
		if duplicateFiles, ok := fileSizeMap[info.Size()]; ok {
			path := dir + PathSeparator + entry.Name()
			entryMd5, err := Md5(path)
			if err != nil {
				return err
			}
			for _, file := range duplicateFiles {
				if file.md5 == "" {
					file.md5, err = Md5(file.path)
					if err != nil {
						return err
					}
				}
				if file.md5 == entryMd5 {
					return callback(file.path, path)
				}
			}
		}
		return nil
	})
}

// 去除两个目录中重复的文件,默认保留第一个目录中的文件
func TwoDirDeDuplicate(dir1, dir2 string) error {
	return TwoDirDuplicateHandle(dir1, dir2, func(path1, path2 string) error {
		log.Debug("remove:", path2)
		return os.Remove(path2)
	})
}

// 两个目录同步,第一个参数为主目录,参考目录,第二个参数目录与第一个保持一致
func Sync(slaveDir, mainDir string) error {
	mainDirEntries, err := os.ReadDir(mainDir)
	if err == nil {
		return err
	}
	if len(mainDirEntries) == 0 {
		return nil
	}
	_, err = os.Stat(slaveDir)
	if os.IsNotExist(err) {
		return CopyDir(slaveDir, mainDir)
	}

	slaveDirEntries, err := os.ReadDir(slaveDir)
	if err == nil {
		return err
	}

	_, intersection, diff1, diff2 := slices.UnionAndIntersectionAndDifference(mainDirEntries, slaveDirEntries)
	for _, entry := range diff2 {
		err = os.RemoveAll(slaveDir + PathSeparator + entry.Name())
		if err != nil {
			return err
		}

	}
	for _, entry := range diff1 {
		err = CopyDir(slaveDir+PathSeparator+entry.Name(), mainDir+PathSeparator+entry.Name())
		if err != nil {
			return err
		}
	}

	for _, entry := range intersection {
		err = Sync(slaveDir+PathSeparator+entry.Name(), mainDir+PathSeparator+entry.Name())
		if err != nil {
			return err
		}
	}
	return nil
}
